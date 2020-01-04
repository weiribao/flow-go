package voter

import (
	"fmt"
	"testing"

	"github.com/dapperlabs/flow-go/model/flow"
	"github.com/dapperlabs/flow-go/module/local"
	protocol "github.com/dapperlabs/flow-go/protocol/badger"
	badger "github.com/dgraph-io/badger/v2"
)

func TestNewVoter(t *testing.T) {
	db, err := badger.Open(badger.DefaultOptions("data").WithLogger(nil))
	if err != nil {
		panic("could not open key-value store")
	}

	state, err := protocol.NewState(db)
	var ids flow.IdentityList
	for _, entry := range []string{"consensus-node1@address1=1000"} {
		id, err := flow.ParseIdentity(entry)
		if err != nil {
			panic("could not parse identity")
		}
		ids = append(ids, id)
	}

	err = state.Mutate().Bootstrap(flow.Genesis(ids))
	if err != nil {
		panic("could not bootstrap protocol State")
	}

	trueID, err := flow.HexStringToIdentifier("node1")
	allIdentities, err := state.Final().Identities()
	fmt.Sprintf("%v", allIdentities)

	id, err := state.Final().Identity(trueID)
	// fnb.MustNot(err).Msg("could not get identity")

	me, err := local.New(id)
	fmt.Sprintf("%v", me)
}
