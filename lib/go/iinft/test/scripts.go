package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/evergreen"
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

// inspectNFTSupplyScript creates a script that reads
// the total supply of tokens in existence
// and makes assertions about the number
func inspectNFTSupplyScript(addrMap map[string]string, tokenContractName string, expectedSupply int) string {
	template := `
		import NonFungibleToken from %s
		import %s from %s

		access(all) fun main() {
			assert(
                %s.totalSupply == UInt64(%d),
                message: "incorrect totalSupply!"
            )
		}
	`

	return fmt.Sprintf(template, addrMap["NonFungibleToken"], tokenContractName, addrMap[tokenContractName], tokenContractName, expectedSupply)
}

// inspectCollectionLenScript creates a script that retrieves an NFT collection
// from storage and tries to borrow a reference for an NFT that it owns.
// If it owns it, it will not fail.
func inspectCollectionLenScript(addrMap map[string]string, userAddr, tokenContractName, publicLocation string, length int) string {
	template := `
		import NonFungibleToken from %s
		import %s from %s

		access(all) fun main() {
			let acct = getAccount(0x%s)
			let collectionRef = acct.capabilities.borrow<&{NonFungibleToken.CollectionPublic}>(%s)
				?? panic("Could not borrow capability from public collection")
			
			if %d != collectionRef.getIDs().length {
				panic("Collection Length is not correct")
			}
		}
	`

	return fmt.Sprintf(template, addrMap["NonFungibleToken"], tokenContractName, addrMap[tokenContractName], userAddr, publicLocation, length)
}

// inspectCollectionScript creates a script that retrieves an NFT collection
// from storage and tries to borrow a reference for an NFT that it owns.
// If it owns it, it will not fail.
func inspectCollectionScript(addrMap map[string]string, userAddr, tokenContractName, publicLocation string, nftID uint64) string {
	template := `
		import NonFungibleToken from %s
		import %s from %s

		access(all) fun main() {
			let acct = getAccount(0x%s)
			let collectionRef = acct.capabilities.borrow<&{NonFungibleToken.CollectionPublic}>(%s)
				?? panic("Could not borrow capability from public collection")
			
			let tokenRef = collectionRef.borrowNFT(UInt64(%d))
		}
	`

	return fmt.Sprintf(template, addrMap["NonFungibleToken"], tokenContractName, addrMap[tokenContractName], userAddr, publicLocation, nftID)
}
