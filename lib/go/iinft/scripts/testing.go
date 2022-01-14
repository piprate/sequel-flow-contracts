package scripts

import (
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/stretchr/testify/require"
)

func FundAccountWithFlow(t *testing.T, se *Engine, receiverAddress flow.Address, amount string) {
	_ = se.NewTransaction("account_fund_flow").
		Argument(cadence.NewAddress(receiverAddress)).
		UFix64Argument(amount).
		SignProposeAndPayAsService().
		Test(t).
		AssertSuccess()
}

func PrepareFUSDMinter(t *testing.T, se *Engine, minterAddress flow.Address) {
	_ = se.NewTransaction("service_setup_fusd_minter").
		SignProposeAndPayAsService().
		Test(t).
		AssertSuccess()

	_ = se.NewTransaction("service_deposit_fusd_minter").
		Argument(cadence.NewAddress(minterAddress)).
		SignProposeAndPayAsService().
		Test(t).
		AssertSuccess()
}

func FundAccountWithFUSD(t *testing.T, se *Engine, receiverAddress flow.Address, amount string) {
	_ = se.NewTransaction("account_fund_fusd").
		Argument(cadence.NewAddress(receiverAddress)).
		UFix64Argument(amount).
		SignProposeAndPayAsService().
		Test(t).
		AssertSuccess()
}

func GetFUSDBalance(t *testing.T, se *Engine, address flow.Address) float64 {
	v, err := se.NewScript("account_balance_fusd").
		Argument(cadence.NewAddress(address)).
		RunReturns()
	require.NoError(t, err)

	return iinft.ToFloat64(v)
}
