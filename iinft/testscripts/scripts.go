package testscripts

import (
	"context"
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/piprate/sequel-flow-contracts/iinft"
	"github.com/piprate/sequel-flow-contracts/iinft/evergreen"
	"github.com/piprate/splash"
	"github.com/stretchr/testify/require"
)

func ConfigureInMemoryEmulator(t *testing.T, client *splash.Connector, adminFlowDeposit string) {
	t.Helper()

	_, err := client.DoNotPrependNetworkToAccountNames().CreateAccountsE(context.Background(), "emulator-account")
	require.NoError(t, err)

	if adminFlowDeposit != "" {
		adminAcct := client.Account("emulator-sequel-admin")

		se, err := iinft.NewTemplateEngine(client)
		require.NoError(t, err)

		FundAccountWithFlow(t, se, adminAcct.Address, adminFlowDeposit)
	}

	err = client.InitializeContractsE(context.Background())
	require.NoError(t, err)
}

func FundAccountWithFlow(t *testing.T, se *splash.TemplateEngine, receiverAddress flow.Address, amount string) {
	t.Helper()

	_ = se.NewTransaction("account_fund_flow").
		Argument(cadence.NewAddress(receiverAddress)).
		UFix64Argument(amount).
		SignProposeAndPayAsService().
		Test(t).
		AssertSuccess()
}

func GetFlowBalance(t *testing.T, se *splash.TemplateEngine, address flow.Address) float64 {
	t.Helper()

	v, err := se.NewScript("account_balance_flow").
		Argument(cadence.NewAddress(address)).
		RunReturns(context.Background())
	require.NoError(t, err)

	return splash.ToFloat64(v)
}

func FundAccountWithExampleToken(t *testing.T, se *splash.TemplateEngine, receiverAddress flow.Address, amount string) {
	t.Helper()

	_ = se.NewTransaction("account_fund_example_ft").
		Argument(cadence.NewAddress(receiverAddress)).
		UFix64Argument(amount).
		SignProposeAndPayAsService().
		Test(t).
		AssertSuccess()
}

func GetExampleTokenBalance(t *testing.T, se *splash.TemplateEngine, address flow.Address) float64 {
	t.Helper()

	v, err := se.NewScript("account_balance_example_ft").
		Argument(cadence.NewAddress(address)).
		RunReturns(context.Background())
	require.NoError(t, err)

	return splash.ToFloat64(v)
}

func SetUpRoyaltyReceivers(t *testing.T, se *splash.TemplateEngine, signAndProposeAs, payAs string, extraTokenNames ...string) {
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

func CreateSealDigitalArtTx(t *testing.T, se *splash.TemplateEngine, client *splash.Connector, metadata *iinft.DigitalArtMetadata,
	profile *evergreen.Profile) splash.FlowTransactionBuilder {
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
