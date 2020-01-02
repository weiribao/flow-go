package voter

import (
	"github.com/dapperlabs/flow-go/engine/consensus/HotStuff/components/voter/voterEvents"
	"github.com/dapperlabs/flow-go/engine/consensus/HotStuff/modules"
	"github.com/dapperlabs/flow-go/engine/consensus/HotStuff/modules/crypto"
	"github.com/dapperlabs/flow-go/engine/consensus/HotStuff/modules/def"
	"github.com/dapperlabs/flow-go/engine/consensus/HotStuff/modules/defConAct"
	"github.com/dapperlabs/flow-go/model/flow"
	"github.com/dapperlabs/flow-go/model/flow/identity"
	"github.com/dapperlabs/flow-go/module"
	"github.com/dapperlabs/flow-go/protocol"
	"go.uber.org/atomic"
)

// Voter contains the logic for deciding whether to vote or not on a safe block and emitting the vote.
// Specifically, Voter processes on OnSafeBlock events from reactor/core/eventprocessor
type Voter struct {
	// enable or disable Voter, e.g. for when catching up blocks
	enabled    atomic.Bool
	view       uint64
	blockKnown map[string]bool
	me         module.Local
	state      protocol.State
	// really should rename modules to prevent confusion with module...
	primaryFetcher modules.PrimaryFetcher
	processor      voterEvents.Processor
	// TODO: Voter needs: SelfID (e.me in Max's Coldstuff) and PrimaryForView (that builds
	// on top of protocol.state ) interfaces
	// BlockProducer needs: SelfID and protocol.state
}

func NewVoter(enabled bool, me module.Local, state protocol.State, primaryFetcher modules.PrimaryFetcher, processor voterEvents.Processor) *Voter {
	return &Voter{*atomic.NewBool(enabled), 0, make(map[string]bool), me, state, primaryFetcher, processor}
}

func (v *Voter) Enable() { v.enabled.Store(true) }

func (v *Voter) Disable() { v.enabled.Store(false) }

// OnSafeBlock votes once Reactor has deemed that a block should be voted on
func (v *Voter) OnSafeBlock(block *def.Block) {
	go v.vote(block)
}

// vote sends a vote to the next primary if Voter is enabled
func (v *Voter) vote(block *def.Block) {
	if !v.enabled.Load() {
		return
	}
	nextPrimary := v.primaryFetcher.ByView(v.view + 1)
	selfIdx := v.getSelfIdx(block.QC.BlockMRH)
	vote := v.generateAndSignVote(block, selfIdx)
	v.processor.OnProducedVote(vote, nextPrimary.Address)
}

func (v *Voter) generateAndSignVote(block *def.Block, selfIdx uint32) *defConAct.Vote {
	voteSig := crypto.SignMsg(block.BlockMRH, selfIdx)
	vote := defConAct.NewVote(block.View, block.BlockMRH, voteSig)

	return vote
}

func (v *Voter) getSelfIdx(blockMRH []byte) uint32 {
	snapshot := v.state.AtHash(blockMRH)
	conIdentities, err := snapshot.Identities(identity.HasRole(flow.RoleConsensus))
	if err != nil {
		// TODO: change logger
		panic("Can't get identity list of consensus nodes")
	}

	selfIdx, err := conIdentities.GetIdx(v.me.NodeID())
	if err != nil {
		// TODO: change logger
		panic("Can't get identity of self from consensus identity list")
	}

	return selfIdx
}

func (v *Voter) OnEnteringView(view uint64) {
	v.view = view
}
