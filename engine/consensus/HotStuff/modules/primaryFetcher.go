package modules

import "github.com/dapperlabs/flow-go/model/flow"

type PrimaryFetcher interface {
	ByView(uint64) flow.Identity
	ByIdentifier([32]byte) flow.Identity
}
