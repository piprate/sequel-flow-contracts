package test

import (
	"fmt"
	"testing"

	"github.com/onflow/cadence"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/splash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarketplace_Integration_ListAndBuyWithFlow(t *testing.T) {
	client, err := splash.NewInMemoryTestConnector("../../../..", true)
	require.NoError(t, err)

	ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := iinft.NewTemplateEngine(client)
	require.NoError(t, err)

	platformAcct := client.Account(platformAccountName)

	// set up seller account

	sellerAcctName := user1AccountName
	sellerAcct := client.Account(sellerAcctName)

	FundAccountWithFlow(t, se, sellerAcct.Address, "10.0")

	_ = se.NewTransaction("account_setup").SignProposeAndPayAs(sellerAcctName).Test(t).AssertSuccess()

	SetUpRoyaltyReceivers(t, se, sellerAcctName, sellerAcctName)

	// set up buyer account

	buyerAcctName := user2AccountName
	buyerAcct := client.Account(buyerAcctName)

	FundAccountWithFlow(t, se, buyerAcct.Address, "10.0")

	_ = se.NewTransaction("account_setup").SignProposeAndPayAs(buyerAcctName).Test(t).AssertSuccess()
	_ = se.NewTransaction("account_setup_flow_token").SignProposeAndPayAs(buyerAcctName).Test(t).AssertSuccess()
	FundAccountWithFlow(t, se, buyerAcct.Address, "1000.0")

	metadata := SampleMetadata(1)
	profile := PrimaryOnlyEvergreenProfile(sellerAcct.Address, platformAcct.Address)

	_ = CreateSealDigitalArtTx(t, se, client, metadata, profile).
		SignProposeAndPayAs(adminAccountName).
		Test(t).
		AssertSuccess()

	res := client.Transaction(se.GetStandardScript("digitalart_mint_edition")).
		SignProposeAndPayAs(adminAccountName).
		StringArgument(metadata.Asset).
		UInt64Argument(1).
		Argument(cadence.Address(sellerAcct.Address)).
		Test(t).
		AssertSuccess()

	nftID := splash.ExtractUInt64ValueFromEvent(res,
		"A.179b6b1cb6755e31.DigitalArt.Minted", "id")

	// Assert that the account's collection is correct
	checkTokenInDigitalArtCollection(t, se, sellerAcct.Address.String(), nftID)
	checkDigitalArtCollectionLen(t, se, sellerAcct.Address.String(), 1)
	checkDigitalArtCollectionLen(t, se, buyerAcct.Address.String(), 0)

	var listingID uint64

	t.Run("Should be able to list an NFT in seller's Storefront", func(t *testing.T) {
		res := se.NewTransaction("marketplace_list").
			SignProposeAndPayAs(sellerAcctName).
			UInt64Argument(nftID).
			UFix64Argument("200.0").
			Argument(cadence.NewAddress(se.ContractAddress("FlowToken"))).
			StringArgument("FlowToken").
			Argument(cadence.NewOptional(cadence.String("link"))).
			Test(t).
			AssertSuccess().
			AssertPartialEvent(splash.NewTestEvent(
				"A.179b6b1cb6755e31.SequelMarketplace.TokenListed",
				map[string]interface{}{
					"metadataLink":     "link",
					"asset":            "did:sequel:asset-id",
					"nftID":            fmt.Sprintf("%d", nftID),
					"nftType":          "A.179b6b1cb6755e31.DigitalArt.NFT",
					"paymentVaultType": "A.0ae53cb6e3f42a79.FlowToken.Vault",
					"payments": []interface{}{
						map[string]interface{}{
							"amount":   "10.00000000",
							"rate":     "0.05000000",
							"receiver": "0xe03daebed8ca0615",
							"role":     "Artist",
						},
						map[string]interface{}{
							"amount":   "190.00000000",
							"rate":     "0.95000000",
							"receiver": "0xe03daebed8ca0615",
							"role":     "Owner",
						},
					},
					"price":             "200.00000000",
					"storefrontAddress": "0xe03daebed8ca0615",
				})).
			AssertPartialEvent(splash.NewTestEvent(
				"A.f8d6e0586b0a20c7.NFTStorefront.ListingAvailable",
				map[string]interface{}{
					"ftVaultType":       "Type\u003cA.0ae53cb6e3f42a79.FlowToken.Vault\u003e()",
					"nftID":             fmt.Sprintf("%d", nftID),
					"nftType":           "Type\u003cA.179b6b1cb6755e31.DigitalArt.NFT\u003e()",
					"price":             "200.00000000",
					"storefrontAddress": "0xe03daebed8ca0615",
				}))

		listingID = splash.ExtractUInt64ValueFromEvent(res,
			"A.179b6b1cb6755e31.SequelMarketplace.TokenListed", "listingID")

		// test listing IDs separately, as they aren't stable
		assert.NotZero(t, listingID)
		assert.Equal(t, listingID, splash.ExtractUInt64ValueFromEvent(res,
			"A.f8d6e0586b0a20c7.NFTStorefront.ListingAvailable", "listingResourceID"))
	})

	t.Run("Should be able to buy an NFT from seller's Storefront", func(t *testing.T) {
		_ = se.NewTransaction("marketplace_buy").
			SignProposeAndPayAs(buyerAcctName).
			UInt64Argument(listingID).
			Argument(cadence.NewAddress(sellerAcct.Address)).
			Argument(cadence.NewAddress(se.ContractAddress("FlowToken"))).
			StringArgument("FlowToken").
			Argument(cadence.NewOptional(cadence.String("link"))).
			Test(t).
			AssertSuccess().
			AssertEmitEvent(splash.NewTestEvent(
				"A.179b6b1cb6755e31.SequelMarketplace.TokenSold",
				map[string]interface{}{
					"listingID":         fmt.Sprintf("%d", listingID),
					"nftID":             fmt.Sprintf("%d", nftID),
					"nftType":           "A.179b6b1cb6755e31.DigitalArt.NFT",
					"paymentVaultType":  "A.0ae53cb6e3f42a79.FlowToken.Vault",
					"price":             "200.00000000",
					"storefrontAddress": "0xe03daebed8ca0615",
					"buyerAddress":      "0x045a1763c93006ca",
					"metadataLink":      "link",
				}))

		// Assert that the account's collection is correct
		checkTokenInDigitalArtCollection(t, se, buyerAcct.Address.String(), 0)
		checkDigitalArtCollectionLen(t, se, buyerAcct.Address.String(), 1)
		checkDigitalArtCollectionLen(t, se, sellerAcct.Address.String(), 0)
	})
}

func TestMarketplace_Integration_ListAndBuyWithExampleToken(t *testing.T) {
	client, err := splash.NewInMemoryTestConnector("../../../..", true)
	require.NoError(t, err)

	ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := iinft.NewTemplateEngine(client)
	require.NoError(t, err)

	platformAcct := client.Account(platformAccountName)

	// set up seller account (seller is the artist)

	sellerAcctName := user1AccountName
	sellerAcct := client.Account(sellerAcctName)

	FundAccountWithFlow(t, se, sellerAcct.Address, "10.0")

	_ = se.NewTransaction("account_setup").SignProposeAndPayAs(sellerAcctName).Test(t).AssertSuccess()

	SetUpRoyaltyReceivers(t, se, sellerAcctName, sellerAcctName, "ExampleToken")

	// set up buyer account

	buyerAcctName := user2AccountName
	buyerAcct := client.Account(buyerAcctName)
	require.NoError(t, err)

	FundAccountWithFlow(t, se, buyerAcct.Address, "10.0")

	_ = se.NewTransaction("account_setup").SignProposeAndPayAs(buyerAcctName).Test(t).AssertSuccess()
	_ = se.NewTransaction("account_setup_example_ft").SignProposeAndPayAs(buyerAcctName).Test(t).AssertSuccess()
	FundAccountWithExampleToken(t, se, buyerAcct.Address, "1000.0")

	metadata := SampleMetadata(1)
	profile := PrimaryOnlyEvergreenProfile(sellerAcct.Address, platformAcct.Address)

	_ = CreateSealDigitalArtTx(t, se, client, metadata, profile).
		SignProposeAndPayAs(adminAccountName).
		Test(t).
		AssertSuccess()

	res := client.Transaction(se.GetStandardScript("digitalart_mint_edition")).
		SignProposeAndPayAs(adminAccountName).
		StringArgument(metadata.Asset).
		UInt64Argument(1).
		Argument(cadence.Address(sellerAcct.Address)).
		Test(t).
		AssertSuccess()

	nftID := splash.ExtractUInt64ValueFromEvent(res,
		"A.179b6b1cb6755e31.DigitalArt.Minted", "id")

	// Assert that the account's collection is correct
	checkTokenInDigitalArtCollection(t, se, sellerAcct.Address.String(), nftID)
	checkDigitalArtCollectionLen(t, se, sellerAcct.Address.String(), 1)
	checkDigitalArtCollectionLen(t, se, buyerAcct.Address.String(), 0)

	var listingID uint64

	t.Run("Should be able to list an NFT in seller's Storefront", func(t *testing.T) {
		res := se.NewTransaction("marketplace_list").
			SignProposeAndPayAs(sellerAcctName).
			UInt64Argument(nftID).
			UFix64Argument("200.0").
			Argument(cadence.NewAddress(se.ContractAddress("ExampleToken"))).
			StringArgument("ExampleToken").
			Argument(cadence.NewOptional(nil)).
			Test(t).
			AssertSuccess().
			AssertPartialEvent(splash.NewTestEvent(
				"A.179b6b1cb6755e31.SequelMarketplace.TokenListed",
				map[string]interface{}{
					"asset":            "did:sequel:asset-id",
					"metadataLink":     "",
					"nftID":            fmt.Sprintf("%d", nftID),
					"nftType":          "A.179b6b1cb6755e31.DigitalArt.NFT",
					"paymentVaultType": "A.f8d6e0586b0a20c7.ExampleToken.Vault",
					"payments": []interface{}{
						map[string]interface{}{
							"amount":   "10.00000000",
							"rate":     "0.05000000",
							"receiver": "0xe03daebed8ca0615",
							"role":     "Artist",
						},
						map[string]interface{}{
							"amount": "190.00000000",

							"rate":     "0.95000000",
							"receiver": "0xe03daebed8ca0615",
							"role":     "Owner",
						},
					},
					"price":             "200.00000000",
					"storefrontAddress": "0xe03daebed8ca0615",
				})).
			AssertPartialEvent(splash.NewTestEvent(
				"A.f8d6e0586b0a20c7.NFTStorefront.ListingAvailable",
				map[string]interface{}{
					"ftVaultType":       "Type\u003cA.f8d6e0586b0a20c7.ExampleToken.Vault\u003e()",
					"nftID":             fmt.Sprintf("%d", nftID),
					"nftType":           "Type\u003cA.179b6b1cb6755e31.DigitalArt.NFT\u003e()",
					"price":             "200.00000000",
					"storefrontAddress": "0xe03daebed8ca0615",
				}))

		listingID = splash.ExtractUInt64ValueFromEvent(res,
			"A.179b6b1cb6755e31.SequelMarketplace.TokenListed", "listingID")

		// test listing IDs separately, as they aren't stable
		assert.NotZero(t, listingID)
		assert.Equal(t, listingID, splash.ExtractUInt64ValueFromEvent(res,
			"A.f8d6e0586b0a20c7.NFTStorefront.ListingAvailable", "listingResourceID"))
	})

	t.Run("Should be able to buy an NFT from seller's Storefront", func(t *testing.T) {
		_ = se.NewTransaction("marketplace_buy").
			SignProposeAndPayAs(buyerAcctName).
			UInt64Argument(listingID).
			Argument(cadence.NewAddress(sellerAcct.Address)).
			Argument(cadence.NewAddress(se.ContractAddress("ExampleToken"))).
			StringArgument("ExampleToken").
			Argument(cadence.NewOptional(nil)).
			Test(t).
			AssertSuccess().
			AssertEmitEvent(splash.NewTestEvent(
				"A.179b6b1cb6755e31.SequelMarketplace.TokenSold",
				map[string]interface{}{
					"listingID":         fmt.Sprintf("%d", listingID),
					"nftID":             fmt.Sprintf("%d", nftID),
					"nftType":           "A.179b6b1cb6755e31.DigitalArt.NFT",
					"paymentVaultType":  "A.f8d6e0586b0a20c7.ExampleToken.Vault",
					"price":             "200.00000000",
					"storefrontAddress": "0xe03daebed8ca0615",
					"buyerAddress":      "0x045a1763c93006ca",
					"metadataLink":      "",
				}))

		// Assert that the account's collection is correct
		checkTokenInDigitalArtCollection(t, se, buyerAcct.Address.String(), 0)
		checkDigitalArtCollectionLen(t, se, buyerAcct.Address.String(), 1)
		checkDigitalArtCollectionLen(t, se, sellerAcct.Address.String(), 0)
	})
}
