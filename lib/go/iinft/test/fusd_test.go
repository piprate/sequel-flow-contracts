package test

import (
	"testing"

	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/scripts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFUSDBalance(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true, true)
	require.NoError(t, err)

	scripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	scripts.PrepareFUSDMinter(t, se, client.Account("emulator-account").Address())

	artistAcctName := "emulator-user1"
	artistAcct := client.Account(artistAcctName)

	assert.Equal(t, 0.0, scripts.GetFUSDBalance(t, se, artistAcct.Address()))

	scripts.FundAccountWithFlow(t, client, artistAcct.Address(), "10.0")

	_ = se.NewTransaction("account_setup_fusd").SignProposeAndPayAs(artistAcctName).Test(t).AssertSuccess()

	assert.Equal(t, 0.0, scripts.GetFUSDBalance(t, se, artistAcct.Address()))

	scripts.FundAccountWithFUSD(t, se, artistAcct.Address(), "123.56")

	assert.Equal(t, 123.56, scripts.GetFUSDBalance(t, se, artistAcct.Address()))
}

func TestSetUpFUSDAccount(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true, true)
	require.NoError(t, err)

	scripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	scripts.PrepareFUSDMinter(t, se, client.Account("emulator-account").Address())

	// set up platform account

	platformAcctName := "emulator-sequel-platform"
	platformAcct := client.Account(platformAcctName)

	scripts.FundAccountWithFlow(t, client, platformAcct.Address(), "10.0")

	artistAcctName := "emulator-user1"

	_ = se.NewTransaction("account_setup_fusd").
		ProposeAs(artistAcctName).
		PayloadSigner(artistAcctName).
		PayAs(platformAcctName).
		Test(t).AssertSuccess()
}
