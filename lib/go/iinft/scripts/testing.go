package scripts

import (
	"bytes"
	"path"
	"strings"
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/evergreen"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/gwtf"
	"github.com/stretchr/testify/require"
)

func ConfigureInMemoryEmulator(t *testing.T, client *gwtf.GoWithTheFlow, adminFlowDeposit string) {
	t.Helper()

	_, err := client.DoNotPrependNetworkToAccountNames().CreateAccountsE("emulator-account")
	require.NoError(t, err)

	if adminFlowDeposit != "" {
		adminAcct := client.Account("emulator-sequel-admin")
		FundAccountWithFlow(t, client, adminAcct.Address(), adminFlowDeposit)
	}

	client.InitializeContracts()
}

func FundAccountWithFlow(t *testing.T, client *gwtf.GoWithTheFlow, receiverAddress flow.Address, amount string) {
	t.Helper()

	contracts := client.State.Contracts().ByNetwork(client.Network)
	addrMap := make(map[string]string)
	for _, contract := range contracts {
		if contract.Alias != "" {
			addrMap[strings.Split(path.Base(contract.Source), ".")[0]] = contract.Alias
		}
	}

	buf := &bytes.Buffer{}
	if err := goTemplates.ExecuteTemplate(buf, "account_fund_flow", addrMap); err != nil {
		panic(err)
	}

	script := buf.String()

	_ = client.Transaction(script).
		Argument(cadence.NewAddress(receiverAddress)).
		UFix64Argument(amount).
		SignProposeAndPayAsService().
		Test(t).
		AssertSuccess()
}

func GetFlowBalance(t *testing.T, se *Engine, address flow.Address) float64 {
	t.Helper()

	v, err := se.NewScript("account_balance_flow").
		Argument(cadence.NewAddress(address)).
		RunReturns()
	require.NoError(t, err)

	return iinft.ToFloat64(v)
}

func PrepareFUSDMinter(t *testing.T, se *Engine, minterAddress flow.Address) {
	t.Helper()

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
	t.Helper()

	_ = se.NewTransaction("account_fund_fusd").
		Argument(cadence.NewAddress(receiverAddress)).
		UFix64Argument(amount).
		SignProposeAndPayAsService().
		Test(t).
		AssertSuccess()
}

func GetFUSDBalance(t *testing.T, se *Engine, address flow.Address) float64 {
	t.Helper()

	v, err := se.NewScript("account_balance_fusd").
		Argument(cadence.NewAddress(address)).
		RunReturns()
	require.NoError(t, err)

	return iinft.ToFloat64(v)
}

func CreateSealDigitalArtTx(t *testing.T, se *Engine, client *gwtf.GoWithTheFlow, metadata *iinft.DigitalArtMetadata,
	profile *evergreen.Profile) gwtf.FlowTransactionBuilder {
	t.Helper()

	profileVal, err := evergreen.ProfileToCadence(profile, flow.HexToAddress(se.WellKnownAddresses()["Evergreen"]))
	require.NoError(t, err)

	tx := client.Transaction(se.GetStandardScript("master_seal")).
		Argument(
			iinft.DigitalArtMetadataToCadence(metadata, flow.HexToAddress(se.WellKnownAddresses()["DigitalArt"])),
		).
		Argument(profileVal)

	return tx
}
