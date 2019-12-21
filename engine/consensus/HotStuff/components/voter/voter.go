package voter

import (
	"github.com/dapperlabs/flow-go/engine/consensus/HotStuff/modules/crypto"
	"github.com/dapperlabs/flow-go/engine/consensus/HotStuff/modules/def"
	"github.com/dapperlabs/flow-go/engine/consensus/HotStuff/modules/defConAct"
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
	// state      protocol.State
	// TODO: Voter needs: SelfID (e.me in Max's Coldstuff) and PrimaryForView (that builds
	// on top of protocol.state ) interfaces
	// BlockProducer needs: SelfID and protocol.state
}

func NewVoter(enabled bool, state protocol.State) *Voter {
	return &Voter{*atomic.NewBool(enabled), 0, make(map[string]bool), state}
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
	nextPrimaryID := v.net.GetNextPrimaryAddress()
	vote := v.GenerateAndSignVote(block, v.conIDTable.GetSelfIdx())
	v.net.SendVote(nextPrimaryAddress, vote)
}

func (v *Voter) GenerateAndSignVote(block *def.Block, selfIdx uint32) *defConAct.Vote {
	voteSig := crypto.SignMsg(block.BlockMRH, selfIdx)
	vote := defConAct.NewVote(block.View, block.BlockMRH, voteSig)

	return vote
}

func (v *Voter) OnEnteringView(view uint64) {
	v.view = view
}
