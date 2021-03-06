// (c) 2019 Dapper Labs - ALL RIGHTS RESERVED

package flow

type Seal struct {
	BlockID       Identifier
	ResultID      Identifier
	InitialState  StateCommitment
	FinalState    StateCommitment
	ServiceEvents []ServiceEvent
}

// TODO need to include service events in hash, omitted for now as they are not
// encodable with RLP
func (s Seal) Body() interface{} {
	return struct {
		BlockID      Identifier
		ResultID     Identifier
		InitialState StateCommitment
		FinalState   StateCommitment
	}{
		BlockID:      s.BlockID,
		ResultID:     s.ResultID,
		InitialState: s.InitialState,
		FinalState:   s.FinalState,
	}
}

func (s Seal) ID() Identifier {
	return MakeID(s.Body())
}

func (s Seal) Checksum() Identifier {
	return MakeID(s)
}
