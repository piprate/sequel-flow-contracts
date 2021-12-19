package scripts

import (
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
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
