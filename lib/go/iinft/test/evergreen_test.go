package test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/evergreen"
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
		description: "Test Role",
		initialSaleCommission: UFix64(0.8),
		secondaryMarketCommission: UFix64(0.05),
		address: addr,
		receiverPath: nil
	)

	assert(role.commissionRate(initialSale: true) == 0.8, message: "wrong commissionRate(true) value")
	assert(role.commissionRate(initialSale: false) == 0.05, message: "wrong commissionRate(false) value")
}
`).
		Argument(cadence.NewAddress(artistAcct.Address)).
		RunReturns(context.Background())
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
		id: "did:sequel:evergreen1",
		description: "Test Profile",
		roles: [
			Evergreen.Role(
				id: "test",
				description: "Test Role",
				initialSaleCommission: UFix64(0.8),
				secondaryMarketCommission: UFix64(0.05),
				address: addr,
				receiverPath: nil
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
		Argument(cadence.NewAddress(artistAcct.Address)).
		RunReturns(context.Background())
	require.NoError(t, err)
}

func TestEvergreen_Profile_buildRoyalties(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true, true)
	require.NoError(t, err)

	scripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	user1Acct := client.Account(user1AccountName)
	user2Acct := client.Account(user2AccountName)

	profile := &evergreen.Profile{
		ID:          "did:sequel:evergreen1",
		Description: "Test Profile",
		Roles: []*evergreen.Role{
			{
				ID:                        "test1",
				Description:               "Test Role 1",
				InitialSaleCommission:     0.8,
				SecondaryMarketCommission: 0.05,
				Address:                   user1Acct.Address,
				ReceiverPath:              "/public/fusdReceiver",
			},
			{
				ID:                        "test2",
				Description:               "Test Role 2",
				InitialSaleCommission:     0.2,
				SecondaryMarketCommission: 0.025,
				Address:                   user2Acct.Address,
			},
		},
	}

	profileVal, err := evergreen.ProfileToCadence(profile, flow.HexToAddress("0x01cf0e2f2f715450"))
	require.NoError(t, err)

	t.Run("Should return no royalties, if no valid receivers", func(t *testing.T) {
		_, err = client.Script(`
import Evergreen from 0x01cf0e2f2f715450

pub fun main(profile: Evergreen.Profile) {
	var royalties = profile.buildRoyalties(defaultReceiverPath: nil)

	assert(royalties.length == 0, message: "Incorrect number of royalties")
}
`).
			Argument(profileVal).
			RunReturns(context.Background())
		require.NoError(t, err)
	})

	scripts.FundAccountWithFlow(t, client, user1Acct.Address, "10.0")
	scripts.FundAccountWithFlow(t, client, user2Acct.Address, "10.0")

	t.Run("if defaultReceiverPath is nil, return royalties with a valid receiver", func(t *testing.T) {
		_ = se.NewTransaction("account_setup_fusd").SignProposeAndPayAs(user1AccountName).
			Test(t).AssertSuccess()

		_, err = client.Script(`
import Evergreen from 0x01cf0e2f2f715450

pub fun main(profile: Evergreen.Profile) {
	var royalties = profile.buildRoyalties(defaultReceiverPath: nil)

	assert(royalties.length == 1, message: "Incorrect number of royalties")
}
`).
			Argument(profileVal).
			RunReturns(context.Background())
		require.NoError(t, err)
	})

	t.Run("if defaultReceiverPath is provided, return royalties with a valid receiver", func(t *testing.T) {
		scripts.FundAccountWithFlow(t, client, user1Acct.Address, "10.0")

		_ = se.NewTransaction("account_setup_fusd").SignProposeAndPayAs(user1AccountName).
			Test(t).AssertSuccess()

		_ = se.NewTransaction("account_setup_fusd").SignProposeAndPayAs(user2AccountName).
			Test(t).AssertSuccess()

		_, err = client.Script(`
import Evergreen from 0x01cf0e2f2f715450

pub fun main(profile: Evergreen.Profile) {
	var royalties = profile.buildRoyalties(defaultReceiverPath: /public/fusdReceiver)

	assert(royalties.length == 2, message: "Incorrect number of royalties")
}
`).
			Argument(profileVal).
			RunReturns(context.Background())
		require.NoError(t, err)
	})
}
