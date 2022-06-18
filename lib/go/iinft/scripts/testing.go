package scripts

import (
	"bytes"
	"fmt"
	"strconv"
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

	err = client.InitializeContractsE()
	require.NoError(t, err)
}

func FundAccountWithFlow(t *testing.T, client *gwtf.GoWithTheFlow, receiverAddress flow.Address, amount string) {
	t.Helper()

	contracts := client.State.Contracts().ByNetwork(client.Network)
	addrMap := make(map[string]string)
	for _, contract := range contracts {
		if contract.Alias != "" {
			addrMap[contract.Name] = contract.Alias
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

func FundAccountWithFlowE(client *gwtf.GoWithTheFlow, receiverAddress flow.Address, amount string) error {
	contracts := client.State.Contracts().ByNetwork(client.Network)
	addrMap := make(map[string]string)
	for _, contract := range contracts {
		if contract.Alias != "" {
			addrMap[contract.Name] = contract.Alias
		}
	}

	buf := &bytes.Buffer{}
	if err := goTemplates.ExecuteTemplate(buf, "account_fund_flow", addrMap); err != nil {
		panic(err)
	}

	script := buf.String()

	_, err := client.Transaction(script).
		Argument(cadence.NewAddress(receiverAddress)).
		UFix64Argument(amount).
		SignProposeAndPayAsService().RunE()

	return err
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

func ExtractStringValueFromEvent(txResult gwtf.TransactionResult, eventName, key string) string {
	for _, e := range txResult.Events {
		if e.Name == eventName {
			v := e.Fields[key]
			if v == nil {
				panic(fmt.Sprintf("key %s not found in %s", key, eventName))
			}
			switch val := v.(type) {
			case string:
				return val
			default:
				panic(fmt.Sprintf("unexpected value type for %s in %s: %T", key, eventName, v))
			}
		}
	}

	return ""
}

func ExtractUInt64ValueFromEvent(txResult gwtf.TransactionResult, eventName, key string) uint64 {
	for _, e := range txResult.Events {
		if e.Name == eventName {
			v := e.Fields[key]
			if v == nil {
				panic(fmt.Sprintf("key %s not found in %s", key, eventName))
			}
			switch val := v.(type) {
			case string:
				res, err := strconv.ParseUint(val, 10, 64)
				if err != nil {
					panic(err)
				}
				return res
			case uint64:
				return val
			default:
				panic(fmt.Sprintf("unexpected value type for %s in %s: %T", key, eventName, v))
			}
		}
	}

	panic(fmt.Sprintf("value not found for %s in %s", key, eventName))
}
