package test

import (
	"testing"

	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/scripts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFUSDBalance(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true)
	require.NoError(t, err)

	client.InitializeContracts().DoNotPrependNetworkToAccountNames().CreateAccounts("emulator-account")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	scripts.PrepareFUSDMinter(t, se, client.Account("emulator-account").Address())

	artistAcctName := "emulator-user1"
	artistAcct := client.Account(artistAcctName)

	assert.Equal(t, 0.0, scripts.GetFUSDBalance(t, se, artistAcct.Address()))

	_ = se.NewTransaction("account_setup_fusd").SignProposeAndPayAs(artistAcctName).Test(t).AssertSuccess()

	assert.Equal(t, 0.0, scripts.GetFUSDBalance(t, se, artistAcct.Address()))
}
