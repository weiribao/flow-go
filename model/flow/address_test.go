package flow

import (
	"encoding/json"
	"math/bits"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type addressWrapper struct {
	Address Address
}

func TestAddressJSON(t *testing.T) {
	addr := Mainnet.Chain().ServiceAddress()
	data, err := json.Marshal(addressWrapper{Address: addr})
	require.Nil(t, err)

	t.Log(string(data))

	var out addressWrapper
	err = json.Unmarshal(data, &out)
	require.Nil(t, err)
	assert.Equal(t, addr, out.Address)
}

func TestShort(t *testing.T) {
	type testcase struct {
		addr     Address
		expected string
	}

	cases := []testcase{
		{
			addr:     Mainnet.Chain().ServiceAddress(),
			expected: Mainnet.Chain().ServiceAddress().Hex(),
		},
		{
			addr:     HexToAddress("0000000002"),
			expected: "02",
		},
		{
			addr:     HexToAddress("1f10"),
			expected: "1f10",
		},
		{
			addr:     HexToAddress("0f10"),
			expected: "0f10",
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.addr.Short(), c.expected)
	}
}

func TestFlowAddressConstants(t *testing.T) {
	// check n and k fit in 8 and 6 bytes
	assert.LessOrEqual(t, linearCodeN, 8*8)
	assert.LessOrEqual(t, linearCodeK, 6*8)

	// Test addresses for all type of networks
	chainIDs := []ChainID{
		Mainnet,
		Testnet,
		Emulator,
	}

	for _, chainID := range chainIDs {

		chain := chainID.Chain()
		if chainID != Emulator {
			// check the Zero and Root constants
			expected := uint64ToAddress(uint64(chainID.getNetworkType()))

			assert.Equal(t, chain.ZeroAddress(), expected)
			expected = uint64ToAddress(generatorMatrixRows[0] ^ uint64(chainID.getNetworkType()))
			assert.Equal(t, chain.ServiceAddress(), expected)
		}

		// check the transition from account zero to root
		state := chain.NewAddressGenerator()
		address, err := state.NextAddress()
		require.NoError(t, err)
		assert.Equal(t, address, chain.ServiceAddress())

		// check high state values: generation should fail for high value states
		state = chain.newAddressGeneratorAtState(maxState)
		_, err = state.NextAddress()
		assert.NoError(t, err)
		_, err = state.NextAddress()
		assert.Error(t, err)

		// check zeroAddress() is an invalid address
		z := chain.ZeroAddress()
		check := chain.IsValid(z)
		assert.False(t, check, "should be invalid")
	}
}

const invalidCodeWord = uint64(0xab2ae42382900010)

func TestAddressGeneration(t *testing.T) {
	// seed random generator
	rand.Seed(time.Now().UnixNano())

	// loops in each test
	const loop = 50

	// Test addresses for all type of networks
	chainIDs := []ChainID{
		Mainnet,
		Testnet,
		Emulator,
	}

	for _, chainID := range chainIDs {

		chain := chainID.Chain()

		// sanity check of NextAddress function consistency
		state := chain.NewAddressGenerator()
		expectedState := uint64(0)
		for i := 0; i < loop; i++ {
			address, err := state.NextAddress()
			require.NoError(t, err)
			expectedState++
			expectedAddress := chain.newAddressGeneratorAtState(expectedState).CurrentAddress()
			assert.Equal(t, address, expectedAddress)
			require.Equal(t, state.CurrentAddress(), expectedAddress)
		}

		// sanity check of addresses weights in Flow.
		// All addresses hamming weights must be less than d.
		// this is only a sanity check of the implementation and not an exhaustive proof
		if chainID == Mainnet {
			r := uint64(rand.Intn(maxState - loop))
			state = chain.newAddressGeneratorAtState(r)
			for i := 0; i < loop; i++ {
				address, err := state.NextAddress()
				require.NoError(t, err)
				weight := bits.OnesCount64(address.uint64())
				assert.LessOrEqual(t, linearCodeD, weight)
			}
		}

		if chainID != Emulator {

			// sanity check of address distances.
			// All distances between any two addresses must be less than d.
			// this is only a sanity check of the implementation and not an exhaustive proof
			r := uint64(rand.Intn(maxState - loop - 1))
			state = chain.newAddressGeneratorAtState(r)
			refAddress, err := state.NextAddress()
			require.NoError(t, err)
			for i := 0; i < loop; i++ {
				address, err := state.NextAddress()
				require.NoError(t, err)
				distance := bits.OnesCount64(address.uint64() ^ refAddress.uint64())
				assert.LessOrEqual(t, linearCodeD, distance)
			}

		}

		// sanity check of valid account addresses.
		// All valid addresses must pass IsValid.
		r := uint64(rand.Intn(maxState - loop))
		state = chain.newAddressGeneratorAtState(r)
		for i := 0; i < loop; i++ {
			address, err := state.NextAddress()
			require.NoError(t, err)
			check := chain.IsValid(address)
			assert.True(t, check, "account address format should be valid")
		}

		// sanity check of invalid account addresses.
		// All invalid addresses must fail IsValid.
		invalidAddress := uint64ToAddress(invalidCodeWord)
		check := chain.IsValid(invalidAddress)
		assert.False(t, check, "account address format should be invalid")
		r = uint64(rand.Intn(maxState - loop))
		state = chain.newAddressGeneratorAtState(r)
		for i := 0; i < loop; i++ {
			address, err := state.NextAddress()
			require.NoError(t, err)
			invalidAddress = uint64ToAddress(address.uint64() ^ invalidCodeWord)
			check := chain.IsValid(invalidAddress)
			assert.False(t, check, "account address format should be invalid")
		}
	}
}

func TestAddressesIntersection(t *testing.T) {
	// seed random generator
	rand.Seed(time.Now().UnixNano())

	// loops in each test
	const loop = 50

	// Test addresses for all type of networks
	chainIDs := []ChainID{
		Testnet,
		Emulator,
	}

	for _, chainID := range chainIDs {

		chain := chainID.Chain()

		// All valid test addresses must fail Flow Mainnet check
		r := uint64(rand.Intn(maxState - loop))
		state := chain.newAddressGeneratorAtState(r)
		for i := 0; i < loop; i++ {
			address, err := state.NextAddress()
			require.NoError(t, err)
			check := Mainnet.Chain().IsValid(address)
			assert.False(t, check, "test account address format should be invalid in Flow")
			sameChainCheck := chain.IsValid(address)
			require.True(t, sameChainCheck)
		}

		// sanity check: mainnet addresses must fail the test check
		r = uint64(rand.Intn(maxState - loop))
		for i := 0; i < loop; i++ {
			invalidAddress, err := Mainnet.Chain().newAddressGeneratorAtState(r).NextAddress()
			require.NoError(t, err)
			check := chain.IsValid(invalidAddress)
			assert.False(t, check, "account address format should be invalid")
		}

		// sanity check of invalid account addresses in all networks
		require.NotEqual(t, invalidCodeWord, uint64(0))
		invalidAddress := uint64ToAddress(invalidCodeWord)
		check := chain.IsValid(invalidAddress)
		assert.False(t, check, "account address format should be invalid")
		r = uint64(rand.Intn(maxState - loop))

		state = chain.newAddressGeneratorAtState(r)
		for i := 0; i < loop; i++ {
			address, err := state.NextAddress()
			require.NoError(t, err)
			invalidAddress = uint64ToAddress(address.uint64() ^ invalidCodeWord)
			// must fail test network check
			check = chain.IsValid(invalidAddress)
			assert.False(t, check, "account address format should be invalid")
			// must fail mainnet check
			check := Mainnet.Chain().IsValid(invalidAddress)
			assert.False(t, check, "account address format should be invalid")
		}
	}
}

func TestUint48(t *testing.T) {
	// seed random generator
	rand.Seed(time.Now().UnixNano())

	const loop = 50
	// test consistensy of putUint48 and uint48
	for i := 0; i < loop; i++ {
		r := uint64(rand.Intn(1 << linearCodeK))
		b := make([]byte, addressStateLength)
		putUint48(b, r)
		res := uint48(b)
		assert.Equal(t, r, res)
	}
}
