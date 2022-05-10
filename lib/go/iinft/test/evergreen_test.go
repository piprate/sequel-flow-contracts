package test

import (
	"os"
	"testing"
	"time"

	"github.com/onflow/cadence"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/scripts"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp})
}

func TestEvergreen_Role_commissionRate(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true, true)
	require.NoError(t, err)

	scripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	artistAcct := client.Account(user1AccountName)

	_, err = client.Script(`
import Evergreen from 0x01cf0e2f2f715450

pub fun main(addr: Address) {

	var role = Evergreen.Role(
		id: "test",
		initialSaleCommission: UFix64(0.8),
		secondaryMarketCommission: UFix64(0.05),
		address: addr
	)

	assert(role.commissionRate(initialSale: true) == 0.8, message: "wrong commissionRate(true) value")
	assert(role.commissionRate(initialSale: false) == 0.05, message: "wrong commissionRate(false) value")
}
`).
		Argument(cadence.NewAddress(artistAcct.Address())).
		RunReturns()
	require.NoError(t, err)
}

func TestEvergreen_Profile_getRole(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true, true)
	require.NoError(t, err)

	scripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	artistAcct := client.Account(user1AccountName)

	_, err = client.Script(`
import Evergreen from 0x01cf0e2f2f715450

pub fun main(addr: Address) {

	var profile = Evergreen.Profile(
		id: 1,
		roles: [
			Evergreen.Role(
				id: "test",
				initialSaleCommission: UFix64(0.8),
				secondaryMarketCommission: UFix64(0.05),
				address: addr
			)
		]
	)

	// getRole should return a role, if exists

	var role = profile.getRole(id: "test")

	assert(role != nil, message: "Role not found")
	assert(role!.id == "test", message: "Wrong role ID")

	role = profile.getRole(id: "bad")

	assert(role == nil, message: "Non-existent role found")
}
`).
		Argument(cadence.NewAddress(artistAcct.Address())).
		RunReturns()
	require.NoError(t, err)
}
