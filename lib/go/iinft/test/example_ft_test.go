package test

import (
	"testing"

	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/splash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetExampleTokenBalance(t *testing.T) {
	client, err := splash.NewInMemoryTestConnector("../../../..", true)
	require.NoError(t, err)

	ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := iinft.NewTemplateEngine(client)
	require.NoError(t, err)

	artistAcctName := user1AccountName
	artistAcct := client.Account(artistAcctName)

	assert.Equal(t, 0.0, GetExampleTokenBalance(t, se, artistAcct.Address))

	FundAccountWithFlow(t, se, artistAcct.Address, "10.0")

	_ = se.NewTransaction("account_setup_example_ft").SignProposeAndPayAs(artistAcctName).Test(t).AssertSuccess()

	assert.Equal(t, 0.0, GetExampleTokenBalance(t, se, artistAcct.Address))

	FundAccountWithExampleToken(t, se, artistAcct.Address, "123.56")

	assert.Equal(t, 123.56, GetExampleTokenBalance(t, se, artistAcct.Address))
}

func TestSetUpExampleTokenAccount(t *testing.T) {
	client, err := splash.NewInMemoryTestConnector("../../../..", true)
	require.NoError(t, err)

	ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := iinft.NewTemplateEngine(client)
	require.NoError(t, err)

	// set up platform account

	platformAcctName := "emulator-sequel-platform"
	platformAcct := client.Account(platformAcctName)

	FundAccountWithFlow(t, se, platformAcct.Address, "10.0")

	artistAcctName := user1AccountName

	_ = se.NewTransaction("account_setup_example_ft").
		ProposeAs(artistAcctName).
		PayloadSigner(artistAcctName).
		PayAs(platformAcctName).
		Test(t).AssertSuccess()
}

func TestAddExampleTokenAsRoyaltyReceiver(t *testing.T) {
	client, err := splash.NewInMemoryTestConnector("../../../..", true)
	require.NoError(t, err)

	ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := iinft.NewTemplateEngine(client)
	require.NoError(t, err)

	SetUpRoyaltyReceivers(t, se, user2AccountName, adminAccountName, "ExampleToken")
}
