package test

import (
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/evergreen"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/gwtf"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/scripts"
	"github.com/stretchr/testify/require"
)

func buildTestProfile(artist flow.Address) *evergreen.Profile {
	return &evergreen.Profile{
		ID: 1,
		Roles: []*evergreen.Role{
			{
				Role:                      evergreen.RoleArtist,
				InitialSaleCommission:     0.8,
				SecondaryMarketCommission: 0.05,
				Address:                   artist,
			},
		},
	}
}

func buildTestMetadata(maxEdition uint64) *iinft.Metadata {
	return &iinft.Metadata{
		MetadataLink:       "QmMetadata",
		Name:               "Pure Art",
		Artist:             "did:sequel:artist",
		Description:        "Digital art in its purest form",
		Type:               "Image",
		ContentLink:        "QmContent",
		ContentPreviewLink: "QmPreview",
		Mimetype:           "image/jpeg",
		MaxEdition:         maxEdition,
		Asset:              "did:sequel:asset-id",
		Record:             "record-id",
		AssetHead:          "asset-head-id",
	}
}

func TestMarketplace_ListAndBuyWithFlow(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true)
	require.NoError(t, err)

	client.CreateAccounts("emulator-account").InitializeContracts().DoNotPrependNetworkToAccountNames()

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

	profile := buildTestProfile(sellerAcct.Address())
	metadata := buildTestMetadata(1)

	_ = scripts.CreateSealDigitalArtTx(se, client, metadata, profile).
		SignProposeAndPayAs(sequelAccount).
		Test(t).
		AssertSuccess()

	_ = client.Transaction(se.GetStandardScript("digitalart_mint_edition")).
		SignProposeAndPayAs(sequelAccount).
		StringArgument(metadata.Asset).
		UInt64Argument(1).
		Argument(cadence.Address(sellerAcct.Address())).
		Test(t).
		AssertSuccess()

	var nftID uint64 = 0

	// Assert that the account's collection is correct
	checkTokenInDigitalArtCollection(t, se, sellerAcct.Address().String(), nftID)
	checkDigitalArtCollectionLen(t, se, sellerAcct.Address().String(), 1)
	checkDigitalArtCollectionLen(t, se, buyerAcct.Address().String(), 0)

	t.Run("Should be able to list an NFT in seller's Storefront", func(t *testing.T) {
		_ = se.NewTransaction("marketplace_list_flow").
			SignProposeAndPayAs(sellerAcctName).
			UInt64Argument(nftID).
			UFix64Argument("200.0").
			Argument(cadence.NewOptional(cadence.String("link"))).
			Test(t).
			AssertSuccess().
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.01cf0e2f2f715450.SequelMarketplace.TokenListed",
				map[string]interface{}{
					"listingID":        "50",
					"metadataLink":     "link",
					"asset":            "did:sequel:asset-id",
					"nftID":            "0",
					"nftType":          "A.01cf0e2f2f715450.DigitalArt.NFT",
					"paymentVaultType": "A.0ae53cb6e3f42a79.FlowToken.Vault",
					"payments": []interface{}{
						map[string]interface{}{
							"amount":   "10.00000000",
							"rate":     "0.05000000",
							"receiver": "0xf3fcd2c1a78f5eee",
							"role":     "Artist",
						},
						map[string]interface{}{
							"amount":   "190.00000000",
							"rate":     "0.95000000",
							"receiver": "0xf3fcd2c1a78f5eee",
							"role":     "Owner",
						},
					},
					"price":             "200.00000000",
					"storefrontAddress": "0xf3fcd2c1a78f5eee",
				})).
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.f8d6e0586b0a20c7.NFTStorefront.ListingAvailable",
				map[string]interface{}{
					"ftVaultType":       "Type\u003cA.0ae53cb6e3f42a79.FlowToken.Vault\u003e()",
					"listingResourceID": "50",
					"nftID":             "0",
					"nftType":           "Type\u003cA.01cf0e2f2f715450.DigitalArt.NFT\u003e()",
					"price":             "200.00000000",
					"storefrontAddress": "0xf3fcd2c1a78f5eee", // seller's address without leading 0
				}))
	})

	t.Run("Should be able to buy an NFT from seller's Storefront", func(t *testing.T) {
		_ = se.NewTransaction("marketplace_buy_flow").
			SignProposeAndPayAs(buyerAcctName).
			UInt64Argument(50).
			Argument(cadence.NewAddress(sellerAcct.Address())).
			Argument(cadence.NewOptional(cadence.String("link"))).
			Test(t).
			AssertSuccess().
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.01cf0e2f2f715450.SequelMarketplace.TokenSold",
				map[string]interface{}{
					"listingID":         "50",
					"nftID":             "0",
					"nftType":           "A.01cf0e2f2f715450.DigitalArt.NFT",
					"paymentVaultType":  "A.0ae53cb6e3f42a79.FlowToken.Vault",
					"price":             "200.00000000",
					"storefrontAddress": "0xf3fcd2c1a78f5eee",
					"buyerAddress":      "0xe03daebed8ca0615",
					"metadataLink":      "link",
				}))

		// Assert that the account's collection is correct
		checkTokenInDigitalArtCollection(t, se, buyerAcct.Address().String(), 0)
		checkDigitalArtCollectionLen(t, se, buyerAcct.Address().String(), 1)
		checkDigitalArtCollectionLen(t, se, sellerAcct.Address().String(), 0)
	})
}

func TestMarketplace_ListAndBuyWithFUSD(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true)
	require.NoError(t, err)

	client.CreateAccounts("emulator-account").InitializeContracts().DoNotPrependNetworkToAccountNames()

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

	profile := buildTestProfile(sellerAcct.Address())
	metadata := buildTestMetadata(1)

	_ = scripts.CreateSealDigitalArtTx(se, client, metadata, profile).
		SignProposeAndPayAs(sequelAccount).
		Test(t).
		AssertSuccess()

	_ = client.Transaction(se.GetStandardScript("digitalart_mint_edition")).
		SignProposeAndPayAs(sequelAccount).
		StringArgument(metadata.Asset).
		UInt64Argument(1).
		Argument(cadence.Address(sellerAcct.Address())).
		Test(t).
		AssertSuccess()

	var nftID uint64 = 0

	// Assert that the account's collection is correct
	checkTokenInDigitalArtCollection(t, se, sellerAcct.Address().String(), nftID)
	checkDigitalArtCollectionLen(t, se, sellerAcct.Address().String(), 1)
	checkDigitalArtCollectionLen(t, se, buyerAcct.Address().String(), 0)

	t.Run("Should be able to list an NFT in seller's Storefront", func(t *testing.T) {
		_ = se.NewTransaction("marketplace_list_fusd").
			SignProposeAndPayAs(sellerAcctName).
			UInt64Argument(nftID).
			UFix64Argument("200.0").
			Argument(cadence.NewOptional(nil)).
			Test(t).
			AssertSuccess().
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.f8d6e0586b0a20c7.NFTStorefront.StorefrontInitialized",
				map[string]interface{}{
					"storefrontResourceID": "52",
				})).
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.01cf0e2f2f715450.SequelMarketplace.TokenListed",
				map[string]interface{}{
					"listingID":        "53",
					"asset":            "did:sequel:asset-id",
					"metadataLink":     "",
					"nftID":            "0",
					"nftType":          "A.01cf0e2f2f715450.DigitalArt.NFT",
					"paymentVaultType": "A.f8d6e0586b0a20c7.FUSD.Vault",
					"payments": []interface{}{
						map[string]interface{}{
							"amount":   "10.00000000",
							"rate":     "0.05000000",
							"receiver": "0xf3fcd2c1a78f5eee",
							"role":     "Artist",
						},
						map[string]interface{}{
							"amount": "190.00000000",

							"rate":     "0.95000000",
							"receiver": "0xf3fcd2c1a78f5eee",
							"role":     "Owner",
						},
					},
					"price":             "200.00000000",
					"storefrontAddress": "0xf3fcd2c1a78f5eee",
				})).
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.f8d6e0586b0a20c7.NFTStorefront.ListingAvailable",
				map[string]interface{}{
					"ftVaultType":       "Type\u003cA.f8d6e0586b0a20c7.FUSD.Vault\u003e()",
					"listingResourceID": "53",
					"nftID":             "0",
					"nftType":           "Type\u003cA.01cf0e2f2f715450.DigitalArt.NFT\u003e()",
					"price":             "200.00000000",
					"storefrontAddress": "0xf3fcd2c1a78f5eee", // seller's address without leading 0
				}))
	})

	t.Run("Should be able to buy an NFT from seller's Storefront", func(t *testing.T) {
		_ = se.NewTransaction("marketplace_buy_fusd").
			SignProposeAndPayAs(buyerAcctName).
			UInt64Argument(53).
			Argument(cadence.NewAddress(sellerAcct.Address())).
			Argument(cadence.NewOptional(nil)).
			Test(t).
			AssertSuccess().
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.01cf0e2f2f715450.SequelMarketplace.TokenSold",
				map[string]interface{}{
					"listingID":         "53",
					"nftID":             "0",
					"nftType":           "A.01cf0e2f2f715450.DigitalArt.NFT",
					"paymentVaultType":  "A.f8d6e0586b0a20c7.FUSD.Vault",
					"price":             "200.00000000",
					"storefrontAddress": "0xf3fcd2c1a78f5eee",
					"buyerAddress":      "0xe03daebed8ca0615",
					"metadataLink":      "",
				}))

		// Assert that the account's collection is correct
		checkTokenInDigitalArtCollection(t, se, buyerAcct.Address().String(), 0)
		checkDigitalArtCollectionLen(t, se, buyerAcct.Address().String(), 1)
		checkDigitalArtCollectionLen(t, se, sellerAcct.Address().String(), 0)
	})
}
