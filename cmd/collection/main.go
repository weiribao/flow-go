package main

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"github.com/onflow/flow-go/cmd"
	"github.com/onflow/flow-go/consensus"
	"github.com/onflow/flow-go/consensus/hotstuff/committee"
	"github.com/onflow/flow-go/consensus/hotstuff/committee/leader"
	"github.com/onflow/flow-go/consensus/hotstuff/notifications"
	"github.com/onflow/flow-go/consensus/hotstuff/pacemaker/timeout"
	"github.com/onflow/flow-go/consensus/hotstuff/verification"
	recovery "github.com/onflow/flow-go/consensus/recovery/protocol"
	"github.com/onflow/flow-go/engine"
	"github.com/onflow/flow-go/engine/collection/epochmgr"
	"github.com/onflow/flow-go/engine/collection/epochmgr/factories"
	"github.com/onflow/flow-go/engine/collection/ingest"
	"github.com/onflow/flow-go/engine/collection/pusher"
	followereng "github.com/onflow/flow-go/engine/common/follower"
	"github.com/onflow/flow-go/engine/common/provider"
	consync "github.com/onflow/flow-go/engine/common/synchronization"
	"github.com/onflow/flow-go/model/encoding"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow-go/model/flow/filter"
	"github.com/onflow/flow-go/module"
	"github.com/onflow/flow-go/module/buffer"
	builder "github.com/onflow/flow-go/module/builder/collection"
	"github.com/onflow/flow-go/module/epochs"
	confinalizer "github.com/onflow/flow-go/module/finalizer/consensus"
	"github.com/onflow/flow-go/module/ingress"
	"github.com/onflow/flow-go/module/mempool"
	"github.com/onflow/flow-go/module/mempool/stdmap"
	"github.com/onflow/flow-go/module/metrics"
	"github.com/onflow/flow-go/module/signature"
	"github.com/onflow/flow-go/module/synchronization"
	storagekv "github.com/onflow/flow-go/storage/badger"
)

func main() {

	var (
		txLimit                                uint
		maxCollectionSize                      uint
		builderExpiryBuffer                    uint
		hotstuffTimeout                        time.Duration
		hotstuffMinTimeout                     time.Duration
		hotstuffTimeoutIncreaseFactor          float64
		hotstuffTimeoutDecreaseFactor          float64
		hotstuffTimeoutVoteAggregationFraction float64
		blockRateDelay                         time.Duration

		ingestConf  ingest.Config
		ingressConf ingress.Config

		pool           mempool.Transactions  // shared tx pool
		followerBuffer *buffer.PendingBlocks // pending block cache for follower

		push              *pusher.Engine
		ing               *ingest.Engine
		mainChainSyncCore *synchronization.Core
		followerEng       *followereng.Engine
		colMetrics        module.CollectionMetrics
		err               error
	)

	cmd.FlowNode(flow.RoleCollection.String()).
		ExtraFlags(func(flags *pflag.FlagSet) {
			flags.UintVar(&txLimit, "tx-limit", 50000,
				"maximum number of transactions in the memory pool")
			flags.StringVarP(&ingressConf.ListenAddr, "ingress-addr", "i", "localhost:9000",
				"the address the ingress server listens on")
			flags.Uint64Var(&ingestConf.MaxGasLimit, "ingest-max-gas-limit", flow.DefaultMaxGasLimit,
				"maximum per-transaction gas limit")
			flags.BoolVar(&ingestConf.CheckScriptsParse, "ingest-check-scripts-parse", true,
				"whether we check that inbound transactions are parse-able")
			flags.BoolVar(&ingestConf.AllowUnknownReference, "ingest-allow-unknown-reference", true,
				"whether we ingest transactions referencing an unknown block")
			flags.UintVar(&ingestConf.ExpiryBuffer, "ingest-expiry-buffer", 30,
				"expiry buffer for inbound transactions")
			flags.UintVar(&ingestConf.PropagationRedundancy, "ingest-tx-propagation-redundancy", 10,
				"how many additional cluster members we propagate transactions to")
			flags.UintVar(&builderExpiryBuffer, "builder-expiry-buffer", 25,
				"expiry buffer for transactions in proposed collections")
			flags.UintVar(&maxCollectionSize, "builder-max-collection-size", 200,
				"maximum number of transactions in proposed collections")
			flags.DurationVar(&hotstuffTimeout, "hotstuff-timeout", 60*time.Second,
				"the initial timeout for the hotstuff pacemaker")
			flags.DurationVar(&hotstuffMinTimeout, "hotstuff-min-timeout", 2500*time.Millisecond,
				"the lower timeout bound for the hotstuff pacemaker")
			flags.Float64Var(&hotstuffTimeoutIncreaseFactor, "hotstuff-timeout-increase-factor",
				timeout.DefaultConfig.TimeoutIncrease,
				"multiplicative increase of timeout value in case of time out event")
			flags.Float64Var(&hotstuffTimeoutDecreaseFactor, "hotstuff-timeout-decrease-factor",
				timeout.DefaultConfig.TimeoutDecrease,
				"multiplicative decrease of timeout value in case of progress")
			flags.Float64Var(&hotstuffTimeoutVoteAggregationFraction, "hotstuff-timeout-vote-aggregation-fraction",
				timeout.DefaultConfig.VoteAggregationTimeoutFraction,
				"additional fraction of replica timeout that the primary will wait for votes")
			flags.DurationVar(&blockRateDelay, "block-rate-delay", 250*time.Millisecond,
				"the delay to broadcast block proposal in order to control block production rate")
		}).
		Module("transactions mempool", func(node *cmd.FlowNodeBuilder) error {
			pool, err = stdmap.NewTransactions(txLimit)
			return err
		}).
		Module("pending block cache", func(node *cmd.FlowNodeBuilder) error {
			followerBuffer = buffer.NewPendingBlocks()
			return nil
		}).
		Module("metrics", func(node *cmd.FlowNodeBuilder) error {
			colMetrics = metrics.NewCollectionCollector(node.Tracer)
			return nil
		}).
		Module("main chain sync core", func(node *cmd.FlowNodeBuilder) error {
			mainChainSyncCore, err = synchronization.New(node.Logger, synchronization.DefaultConfig())
			return err
		}).
		Component("follower engine", func(node *cmd.FlowNodeBuilder) (module.ReadyDoneAware, error) {

			// initialize cleaner for DB
			cleaner := storagekv.NewCleaner(node.Logger, node.DB, metrics.NewCleanerCollector(), flow.DefaultValueLogGCFrequency)

			// create a finalizer that will handling updating the protocol
			// state when the follower detects newly finalized blocks
			finalizer := confinalizer.NewFinalizer(node.DB, node.Storage.Headers, node.State)

			// initialize the staking & beacon verifiers, signature joiner
			staking := signature.NewAggregationVerifier(encoding.ConsensusVoteTag)
			beacon := signature.NewThresholdVerifier(encoding.RandomBeaconTag)
			merger := signature.NewCombiner()

			// initialize and pre-generate leader selections from the seed
			selection, err := leader.NewSelectionForConsensus(leader.EstimatedSixMonthOfViews, node.RootBlock.Header, node.RootQC, node.State)
			if err != nil {
				return nil, fmt.Errorf("could not create leader selection for main consensus: %w", err)
			}

			// initialize consensus committee's membership state
			// This committee state is for the HotStuff follower, which follows the MAIN CONSENSUS Committee
			// Note: node.Me.NodeID() is not part of the consensus committee
			mainConsensusCommittee, err := committee.NewMainConsensusCommitteeState(node.State, node.Me.NodeID(), selection)
			if err != nil {
				return nil, fmt.Errorf("could not create Committee state for main consensus: %w", err)
			}

			// initialize the verifier for the protocol consensus
			verifier := verification.NewCombinedVerifier(mainConsensusCommittee, staking, beacon, merger)

			// use proper engine for notifier to follower
			notifier := notifications.NewNoopConsumer()

			finalized, pending, err := recovery.FindLatest(node.State, node.Storage.Headers)
			if err != nil {
				return nil, fmt.Errorf("could not find latest finalized block and pending blocks to recover consensus follower: %w", err)
			}

			// creates a consensus follower with noop consumer as the notifier
			followerCore, err := consensus.NewFollower(
				node.Logger,
				mainConsensusCommittee,
				node.Storage.Headers,
				finalizer,
				verifier,
				notifier,
				node.RootBlock.Header,
				node.RootQC,
				finalized,
				pending,
			)
			if err != nil {
				return nil, fmt.Errorf("could not create follower core logic: %w", err)
			}

			followerEng, err = followereng.New(
				node.Logger,
				node.Network,
				node.Me,
				node.Metrics.Engine,
				node.Metrics.Mempool,
				cleaner,
				node.Storage.Headers,
				node.Storage.Payloads,
				node.State,
				followerBuffer,
				followerCore,
				mainChainSyncCore,
			)
			if err != nil {
				return nil, fmt.Errorf("could not create follower engine: %w", err)
			}

			return followerEng, nil
		}).
		Component("main chain sync engine", func(node *cmd.FlowNodeBuilder) (module.ReadyDoneAware, error) {

			// create a block synchronization engine to handle follower getting out of sync
			sync, err := consync.New(
				node.Logger,
				node.Metrics.Engine,
				node.Network,
				node.Me,
				node.State,
				node.Storage.Blocks,
				followerEng,
				mainChainSyncCore,
			)
			if err != nil {
				return nil, fmt.Errorf("could not create synchronization engine: %w", err)
			}

			return sync, nil
		}).
		Component("ingestion engine", func(node *cmd.FlowNodeBuilder) (module.ReadyDoneAware, error) {
			ing, err = ingest.New(
				node.Logger,
				node.Network,
				node.State,
				node.Metrics.Engine,
				colMetrics,
				node.Me,
				pool,
				ingestConf,
			)
			return ing, err
		}).
		Component("transaction ingress server", func(node *cmd.FlowNodeBuilder) (module.ReadyDoneAware, error) {
			server := ingress.New(ingressConf, ing, node.RootChainID)
			return server, nil
		}).
		Component("provider engine", func(node *cmd.FlowNodeBuilder) (module.ReadyDoneAware, error) {
			retrieve := func(collID flow.Identifier) (flow.Entity, error) {
				coll, err := node.Storage.Collections.ByID(collID)
				return coll, err
			}
			return provider.New(node.Logger, node.Metrics.Engine, node.Network, node.Me, node.State,
				engine.ProvideCollections,
				filter.HasRole(flow.RoleAccess, flow.RoleExecution),
				retrieve,
			)
		}).
		Component("pusher engine", func(node *cmd.FlowNodeBuilder) (module.ReadyDoneAware, error) {
			push, err = pusher.New(
				node.Logger,
				node.Network,
				node.State,
				node.Metrics.Engine,
				colMetrics,
				node.Me,
				pool,
				node.Storage.Collections,
				node.Storage.Transactions,
			)
			return push, err
		}).
		// Epoch manager encapsulates and manages epoch-dependent engines as we
		// transition between epochs
		Component("epoch manager", func(node *cmd.FlowNodeBuilder) (module.ReadyDoneAware, error) {

			clusterStateFactory, err := factories.NewClusterStateFactory(node.DB, node.Metrics.Cache)
			if err != nil {
				return nil, err
			}

			builderFactory, err := factories.NewBuilderFactory(
				node.DB,
				node.Storage.Headers,
				node.Tracer,
				colMetrics,
				push,
				builder.WithMaxCollectionSize(maxCollectionSize),
				builder.WithExpiryBuffer(builderExpiryBuffer),
			)
			if err != nil {
				return nil, err
			}

			proposalFactory, err := factories.NewProposalEngineFactory(
				node.Logger,
				node.Network,
				node.Me,
				colMetrics,
				node.Metrics.Engine,
				node.Metrics.Mempool,
				node.State,
				pool,
				node.Storage.Transactions,
			)
			if err != nil {
				return nil, err
			}

			syncFactory, err := factories.NewSyncEngineFactory(
				node.Logger,
				node.Metrics.Engine,
				node.Network,
				node.Me,
				synchronization.DefaultConfig(),
			)
			if err != nil {
				return nil, err
			}

			hotstuffFactory, err := factories.NewHotStuffFactory(
				node.Logger,
				node.Me,
				node.DB,
				node.State,
				consensus.WithBlockRateDelay(blockRateDelay),
				consensus.WithInitialTimeout(hotstuffTimeout),
				consensus.WithMinTimeout(hotstuffMinTimeout),
				consensus.WithVoteAggregationTimeoutFraction(hotstuffTimeoutVoteAggregationFraction),
				consensus.WithTimeoutIncreaseFactor(hotstuffTimeoutIncreaseFactor),
				consensus.WithTimeoutDecreaseFactor(hotstuffTimeoutDecreaseFactor),
			)
			if err != nil {
				return nil, err
			}

			staking := signature.NewAggregationProvider(encoding.CollectorVoteTag, node.Me)
			signer := verification.NewSingleSigner(staking, node.Me.NodeID())
			rootQCVoter := epochs.NewRootQCVoter(
				node.Logger,
				node.Me,
				signer,
				node.State,
				nil, // TODO
			)

			factory := factories.NewEpochComponentsFactory(
				node.Me,
				pool,
				builderFactory,
				clusterStateFactory,
				hotstuffFactory,
				proposalFactory,
				syncFactory,
			)

			manager, err := epochmgr.New(
				node.Logger,
				node.Me,
				node.State,
				rootQCVoter,
				factory,
			)
			if err != nil {
				return nil, fmt.Errorf("could not create epoch manager: %w", err)
			}

			// register the manager for protocol events
			node.ProtocolEvents.AddConsumer(manager)

			return manager, err
		}).
		Run()
}
