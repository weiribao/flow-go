package voterEvents

import "github.com/dapperlabs/flow-go/engine/consensus/HotStuff/modules/defConAct"

type Processor interface {
	OnProducedVote(*defConAct.Vote, string)
}

type SentVoteConsumer interface {
	OnProducedVote(*defConAct.Vote)
}
