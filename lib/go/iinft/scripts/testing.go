package scripts

import (
	"bytes"
	"context"
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

	_, err := client.DoNotPrependNetworkToAccountNames().CreateAccountsE(context.Background(), "emulator-account")
	require.NoError(t, err)

	if adminFlowDeposit != "" {
		adminAcct := client.Account("emulator-sequel-admin")
		FundAccountWithFlow(t, client, adminAcct.Address, adminFlowDeposit)
	}

	err = client.InitializeContractsE(context.Background())
	require.NoError(t, err)
}

func FundAccountWithFlow(t *testing.T, client *gwtf.GoWithTheFlow, receiverAddress flow.Address, amount string) {
	t.Helper()

	contracts := client.State.Contracts()
	addrMap := make(map[string]string)
	networkName := client.Services.Network().Name

	for _, contract := range *contracts {
		for _, alias := range contract.Aliases {
			if alias.Network == networkName {
				addrMap[contract.Name] = alias.Address.HexWithPrefix()
			}
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
	contracts := client.State.Contracts() //.ByNetwork(client.Network)
	addrMap := make(map[string]string)
	networkName := client.Services.Network().Name

	for _, contract := range *contracts {
		for _, alias := range contract.Aliases {
			if alias.Network == networkName {
				addrMap[contract.Name] = alias.Address.HexWithPrefix()
			}
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
		SignProposeAndPayAsService().RunE(context.Background())

	return err
}

func GetFlowBalance(t *testing.T, se *Engine, address flow.Address) float64 {
	t.Helper()

	v, err := se.NewScript("account_balance_flow").
		Argument(cadence.NewAddress(address)).
		RunReturns(context.Background())
	require.NoError(t, err)

	return iinft.ToFloat64(v)
}

func FundAccountWithExampleToken(t *testing.T, se *Engine, receiverAddress flow.Address, amount string) {
	t.Helper()

	_ = se.NewTransaction("account_fund_example_ft").
		Argument(cadence.NewAddress(receiverAddress)).
		UFix64Argument(amount).
		SignProposeAndPayAsService().
		Test(t).
		AssertSuccess()
}

func GetExampleTokenBalance(t *testing.T, se *Engine, address flow.Address) float64 {
	t.Helper()

	v, err := se.NewScript("account_balance_example_ft").
		Argument(cadence.NewAddress(address)).
		RunReturns(context.Background())
	require.NoError(t, err)

	return iinft.ToFloat64(v)
}

func SetUpRoyaltyReceivers(t *testing.T, se *Engine, signAndProposeAs, payAs string, extraTokenNames ...string) {
	t.Helper()

	addresses := make([]cadence.Value, len(extraTokenNames))
	names := make([]cadence.Value, len(extraTokenNames))

	for i, name := range extraTokenNames {
		addresses[i] = cadence.NewAddress(se.ContractAddress(name))
		names[i] = cadence.String(name)
	}

	_ = se.NewTransaction("account_royalty_receiver_setup").
		SignAndProposeAs(signAndProposeAs).
		PayAs(payAs).
		Argument(cadence.NewArray(addresses)).
		Argument(cadence.NewArray(names)).
		Test(t).
		AssertSuccess()
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
