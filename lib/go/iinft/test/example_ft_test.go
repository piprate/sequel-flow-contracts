package test

import (
	"testing"

	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/scripts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetExampleTokenBalance(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true, true)
	require.NoError(t, err)

	scripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	artistAcctName := user1AccountName
	artistAcct := client.Account(artistAcctName)

	assert.Equal(t, 0.0, scripts.GetExampleTokenBalance(t, se, artistAcct.Address))

	scripts.FundAccountWithFlow(t, client, artistAcct.Address, "10.0")

	_ = se.NewTransaction("account_setup_example_ft").SignProposeAndPayAs(artistAcctName).Test(t).AssertSuccess()

	assert.Equal(t, 0.0, scripts.GetExampleTokenBalance(t, se, artistAcct.Address))

	scripts.FundAccountWithExampleToken(t, se, artistAcct.Address, "123.56")

	assert.Equal(t, 123.56, scripts.GetExampleTokenBalance(t, se, artistAcct.Address))
}

func TestSetUpExampleTokenAccount(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true, true)
	require.NoError(t, err)

	scripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	// set up platform account

	platformAcctName := "emulator-sequel-platform"
	platformAcct := client.Account(platformAcctName)

	scripts.FundAccountWithFlow(t, client, platformAcct.Address, "10.0")

	artistAcctName := user1AccountName

	_ = se.NewTransaction("account_setup_example_ft").
		ProposeAs(artistAcctName).
		PayloadSigner(artistAcctName).
		PayAs(platformAcctName).
		Test(t).AssertSuccess()
}

func TestAddExampleTokenAsRoyaltyReceiver(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true, true)
	require.NoError(t, err)

	scripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	scripts.SetUpRoyaltyReceivers(t, se, user2AccountName, adminAccountName, "ExampleToken")
}
