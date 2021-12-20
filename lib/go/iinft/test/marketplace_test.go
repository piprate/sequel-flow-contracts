package test

import (
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/gwtf"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/scripts"
	"github.com/stretchr/testify/require"
)

func buildTestMetadata(artist flow.Address, maxEdition uint64) *iinft.Metadata {
	return &iinft.Metadata{
		MetadataLink:       "QmMetadata",
		Name:               "Pure Art",
		Artist:             "Arty",
		Description:        "Digital art in its purest form",
		Type:               "Image",
		ContentLink:        "QmContent",
		ContentPreviewLink: "QmPreview",
		Mimetype:           "image/jpeg",
		MaxEdition:         maxEdition,
		Asset:              "did:sequel:asset-id",
		Record:             "record-id",
		AssetHead:          "asset-head-id",
		EvergreenProfile: &iinft.EvergreenProfile{
			ID: 1,
			Roles: map[string]*iinft.EvergreenRole{
				iinft.EvergreenRoleArtist: {
					Role:                      iinft.EvergreenRoleArtist,
					InitialSaleCommission:     80.0,
					SecondaryMarketCommission: 5.0,
					Address:                   artist,
				},
			},
		},
	}
}

func TestMarketplace_ListAndBuyWithFlow(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true)
	require.NoError(t, err)

	client.InitializeContracts().DoNotPrependNetworkToAccountNames().CreateAccounts("emulator-account")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	// set up seller account

	sellerAcctName := "emulator-user1"
	sellerAcct := client.Account(sellerAcctName)

	_ = se.NewTransaction("account_setup").SignProposeAndPayAs(sellerAcctName).Test(t).AssertSuccess()
	_ = se.NewTransaction("account_setup_flow_token").SignProposeAndPayAs(sellerAcctName).Test(t).AssertSuccess()

	// set up buyer account

	buyerAcctName := "emulator-user2"
	buyerAcct := client.Account(buyerAcctName)

	_ = se.NewTransaction("account_setup").SignProposeAndPayAs(buyerAcctName).Test(t).AssertSuccess()
	_ = se.NewTransaction("account_setup_flow_token").SignProposeAndPayAs(buyerAcctName).Test(t).AssertSuccess()
	scripts.FundAccountWithFlow(t, se, buyerAcct.Address(), "1000.0")

	metadata := buildTestMetadata(sellerAcct.Address(), 0)

	_ = scripts.CreateMintSingleDigitalArtTx(se.GetStandardScript("digitalart_mint_single"), client, metadata, sellerAcct.Address()).
		SignProposeAndPayAs(sequelAccount).
		Test(t).
		AssertSuccess().
		AssertEmitEvent(gwtf.NewTestEvent("A.f8d6e0586b0a20c7.DigitalArt.Minted", map[string]interface{}{
			"id":      "0",
			"asset":   "did:sequel:asset-id",
			"edition": "1",
		}))

	var nftID uint64 = 0

	// Assert that the account's collection is correct
	checkTokenInDigitalArtCollection(t, se, sellerAcct.Address().String(), nftID)
	checkDigitalArtCollectionLen(t, se, sellerAcct.Address().String(), 1)
	checkDigitalArtCollectionLen(t, se, buyerAcct.Address().String(), 0)

	t.Run("Should be able to list an NFT in seller's Storefront", func(t *testing.T) {
		_ = se.NewTransaction("marketplace_add_flow").
			SignProposeAndPayAs(sellerAcctName).
			UInt64Argument(nftID).
			UFix64Argument("200.0").
			BooleanArgument(true).
			Test(t).
			AssertSuccess().
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.f8d6e0586b0a20c7.NFTStorefront.ListingAvailable",
				map[string]interface{}{
					"ftVaultType":       "Type\u003cA.0ae53cb6e3f42a79.FlowToken.Vault\u003e()",
					"listingResourceID": "44",
					"nftID":             "0",
					"nftType":           "Type\u003cA.f8d6e0586b0a20c7.DigitalArt.NFT\u003e()",
					"price":             "200.00000000",
					"storefrontAddress": "0x1cf0e2f2f715450", // seller's address without leading 0
				}))
	})

	t.Run("Should be able to buy an NFT from seller's Storefront", func(t *testing.T) {
		_ = se.NewTransaction("marketplace_buy_flow").
			SignProposeAndPayAs(buyerAcctName).
			UInt64Argument(44).
			Argument(cadence.NewAddress(sellerAcct.Address())).
			Test(t).
			AssertSuccess()

		// Assert that the account's collection is correct
		checkTokenInDigitalArtCollection(t, se, buyerAcct.Address().String(), 0)
		checkDigitalArtCollectionLen(t, se, buyerAcct.Address().String(), 1)
		checkDigitalArtCollectionLen(t, se, sellerAcct.Address().String(), 0)
	})
}

func TestMarketplace_ListAndBuyWithFUSD(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true)
	require.NoError(t, err)

	client.InitializeContracts().DoNotPrependNetworkToAccountNames().CreateAccounts("emulator-account")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	scripts.PrepareFUSDMinter(t, se, client.Account("emulator-account").Address())

	// set up seller account

	sellerAcctName := "emulator-user1"
	sellerAcct := client.Account(sellerAcctName)

	_ = se.NewTransaction("account_setup").SignProposeAndPayAs(sellerAcctName).Test(t).AssertSuccess()
	_ = se.NewTransaction("account_setup_fusd").SignProposeAndPayAs(sellerAcctName).Test(t).AssertSuccess()

	// set up buyer account

	buyerAcctName := "emulator-user2"
	buyerAcct, err := client.State.Accounts().ByName(buyerAcctName)
	require.NoError(t, err)

	_ = se.NewTransaction("account_setup").SignProposeAndPayAs(buyerAcctName).Test(t).AssertSuccess()
	_ = se.NewTransaction("account_setup_fusd").SignProposeAndPayAs(buyerAcctName).Test(t).AssertSuccess()
	scripts.FundAccountWithFUSD(t, se, buyerAcct.Address(), "1000.0")

	metadata := buildTestMetadata(sellerAcct.Address(), 0)

	_ = scripts.CreateMintSingleDigitalArtTx(se.GetStandardScript("digitalart_mint_single"), client, metadata, sellerAcct.Address()).
		SignProposeAndPayAs(sequelAccount).
		Test(t).
		AssertSuccess().
		AssertEmitEvent(gwtf.NewTestEvent("A.f8d6e0586b0a20c7.DigitalArt.Minted", map[string]interface{}{
			"id":      "0",
			"asset":   "did:sequel:asset-id",
			"edition": "1",
		}))

	var nftID uint64 = 0

	// Assert that the account's collection is correct
	checkTokenInDigitalArtCollection(t, se, sellerAcct.Address().String(), nftID)
	checkDigitalArtCollectionLen(t, se, sellerAcct.Address().String(), 1)
	checkDigitalArtCollectionLen(t, se, buyerAcct.Address().String(), 0)

	t.Run("Should be able to list an NFT in seller's Storefront", func(t *testing.T) {
		_ = se.NewTransaction("marketplace_add_fusd").
			SignProposeAndPayAs(sellerAcctName).
			UInt64Argument(nftID).
			UFix64Argument("200.0").
			BooleanArgument(true).
			Test(t).
			AssertSuccess().
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.f8d6e0586b0a20c7.NFTStorefront.StorefrontInitialized",
				map[string]interface{}{
					"storefrontResourceID": "46",
				})).
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.f8d6e0586b0a20c7.NFTStorefront.ListingAvailable",
				map[string]interface{}{
					"ftVaultType":       "Type\u003cA.f8d6e0586b0a20c7.FUSD.Vault\u003e()",
					"listingResourceID": "47",
					"nftID":             "0",
					"nftType":           "Type\u003cA.f8d6e0586b0a20c7.DigitalArt.NFT\u003e()",
					"price":             "200.00000000",
					"storefrontAddress": "0x1cf0e2f2f715450", // seller's address without leading 0
				}))
	})

	t.Run("Should be able to buy an NFT from seller's Storefront", func(t *testing.T) {
		_ = se.NewTransaction("marketplace_buy_fusd").
			SignProposeAndPayAs(buyerAcctName).
			UInt64Argument(47).
			Argument(cadence.NewAddress(sellerAcct.Address())).
			Test(t).
			AssertSuccess()

		// Assert that the account's collection is correct
		checkTokenInDigitalArtCollection(t, se, buyerAcct.Address().String(), 0)
		checkDigitalArtCollectionLen(t, se, buyerAcct.Address().String(), 1)
		checkDigitalArtCollectionLen(t, se, sellerAcct.Address().String(), 0)
	})
}
