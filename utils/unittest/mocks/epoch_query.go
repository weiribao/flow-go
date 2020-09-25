package mocks

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-go/state/protocol"
)

type EpochQuery struct {
	t         *testing.T
	counter   uint64
	byCounter map[uint64]protocol.Epoch
}

func NewEpochQuery(t *testing.T, counter uint64, epochs ...protocol.Epoch) *EpochQuery {
	mock := &EpochQuery{
		t:         t,
		counter:   counter,
		byCounter: make(map[uint64]protocol.Epoch),
	}

	for _, epoch := range epochs {
		mock.Add(epoch)
	}

	return mock
}

func (mock *EpochQuery) Current() protocol.Epoch {
	return mock.byCounter[mock.counter]
}

func (mock *EpochQuery) Next() protocol.Epoch {
	return mock.byCounter[mock.counter+1]
}

func (mock *EpochQuery) ByCounter(counter uint64) protocol.Epoch {
	return mock.byCounter[counter]
}

func (mock *EpochQuery) Transition() {
	mock.counter++
}

func (mock *EpochQuery) Add(epoch protocol.Epoch) {
	counter, err := epoch.Counter()
	assert.Nil(mock.t, err, "cannot add epoch with invalid counter")
	mock.byCounter[counter] = epoch
}