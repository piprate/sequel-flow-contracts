package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/evergreen"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/gwtf"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/scripts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarketplace_listToken(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true, true)
	require.NoError(t, err)

	scripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	scripts.PrepareFUSDMinter(t, se, client.Account("emulator-account").Address)

	platformAcct := client.Account(platformAccountName)

	// set up seller account

	sellerAcctName := user1AccountName
	sellerAcct := client.Account(sellerAcctName)

	scripts.FundAccountWithFlow(t, client, sellerAcct.Address, "10.0")

	_ = se.NewTransaction("account_setup").SignProposeAndPayAs(sellerAcctName).Test(t).AssertSuccess()

	artistAcct := client.Account(user2AccountName)

	_ = se.NewTransaction("account_royalty_receiver_setup").SignAndProposeAs(user2AccountName).PayAs(adminAccountName).Test(t).AssertSuccess()

	metadata := SampleMetadata(1)
	profile := PrimaryOnlyEvergreenProfile(artistAcct.Address, platformAcct.Address)

	_ = scripts.CreateSealDigitalArtTx(t, se, client, metadata, profile).
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

	nftID := scripts.ExtractUInt64ValueFromEvent(res,
		"A.01cf0e2f2f715450.DigitalArt.Minted", "id")

	// Assert that the account's collection is correct
	checkTokenInDigitalArtCollection(t, se, sellerAcct.Address.String(), nftID)
	checkDigitalArtCollectionLen(t, se, sellerAcct.Address.String(), 1)

	t.Run("Happy path (Flow)", func(t *testing.T) {
		res := se.NewTransaction("marketplace_list_flow").
			SignProposeAndPayAs(sellerAcctName).
			UInt64Argument(nftID).
			UFix64Argument("200.0").
			Argument(cadence.NewOptional(cadence.String("link"))).
			Test(t).
			AssertSuccess().
			AssertPartialEvent(gwtf.NewTestEvent(
				"A.01cf0e2f2f715450.SequelMarketplace.TokenListed",
				map[string]interface{}{
					"asset":            "did:sequel:asset-id",
					"metadataLink":     "link",
					"nftID":            fmt.Sprintf("%d", nftID),
					"nftType":          "A.01cf0e2f2f715450.DigitalArt.NFT",
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
							"receiver": "0xf3fcd2c1a78f5eee",
							"role":     "Owner",
						},
					},
					"price":             "200.00000000",
					"storefrontAddress": "0xf3fcd2c1a78f5eee",
				})).
			AssertPartialEvent(gwtf.NewTestEvent(
				"A.f8d6e0586b0a20c7.NFTStorefront.ListingAvailable",
				map[string]interface{}{
					"ftVaultType":       "Type\u003cA.0ae53cb6e3f42a79.FlowToken.Vault\u003e()",
					"nftID":             fmt.Sprintf("%d", nftID),
					"nftType":           "Type\u003cA.01cf0e2f2f715450.DigitalArt.NFT\u003e()",
					"price":             "200.00000000",
					"storefrontAddress": "0xf3fcd2c1a78f5eee",
				}))

		// test listing IDs separately, as they aren't stable
		assert.NotZero(t, scripts.ExtractUInt64ValueFromEvent(res,
			"A.01cf0e2f2f715450.SequelMarketplace.TokenListed", "listingID"))
		assert.NotZero(t, scripts.ExtractUInt64ValueFromEvent(res,
			"A.f8d6e0586b0a20c7.NFTStorefront.ListingAvailable", "listingResourceID"))
	})

	t.Run("Fail, if seller's receiver is invalid (FUSD)", func(t *testing.T) {
		// Fund with Flow for FUSD setup fees
		scripts.FundAccountWithFlow(t, client, artistAcct.Address, "10.0")

		_ = se.NewTransaction("account_setup_fusd").SignProposeAndPayAs(artistAcct.Name).Test(t).AssertSuccess()

		_ = se.NewTransaction("marketplace_list_fusd").
			SignProposeAndPayAs(sellerAcctName).
			UInt64Argument(nftID).
			UFix64Argument("200.0").
			Argument(cadence.NewOptional(cadence.String("link"))).
			Test(t).
			AssertFailure("missing fungible token receiver")
	})

	t.Run("Succeed, if some receivers are invalid (FUSD)", func(t *testing.T) {
		// Fund with Flow for FUSD setup fees
		scripts.FundAccountWithFlow(t, client, artistAcct.Address, "10.0")

		_ = se.NewTransaction("account_setup_fusd").SignProposeAndPayAs(sellerAcctName).Test(t).AssertSuccess()
		_ = se.NewTransaction("account_setup_fusd").SignProposeAndPayAs(artistAcct.Name).Test(t).AssertSuccess()

		res := se.NewTransaction("marketplace_list_fusd").
			SignProposeAndPayAs(sellerAcctName).
			UInt64Argument(nftID).
			UFix64Argument("200.0").
			Argument(cadence.NewOptional(cadence.String("link"))).
			Test(t).
			AssertSuccess().
			AssertPartialEvent(gwtf.NewTestEvent(
				"A.01cf0e2f2f715450.SequelMarketplace.TokenListed",
				map[string]interface{}{
					"asset":            "did:sequel:asset-id",
					"metadataLink":     "link",
					"nftID":            fmt.Sprintf("%d", nftID),
					"nftType":          "A.01cf0e2f2f715450.DigitalArt.NFT",
					"paymentVaultType": "A.f8d6e0586b0a20c7.FUSD.Vault",
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
							"receiver": "0xf3fcd2c1a78f5eee",
							"role":     "Owner",
						},
					},
					"price":             "200.00000000",
					"storefrontAddress": "0xf3fcd2c1a78f5eee",
				})).
			AssertPartialEvent(gwtf.NewTestEvent(
				"A.f8d6e0586b0a20c7.NFTStorefront.ListingAvailable",
				map[string]interface{}{
					"ftVaultType":       "Type\u003cA.f8d6e0586b0a20c7.FUSD.Vault\u003e()",
					"nftID":             fmt.Sprintf("%d", nftID),
					"nftType":           "Type\u003cA.01cf0e2f2f715450.DigitalArt.NFT\u003e()",
					"price":             "200.00000000",
					"storefrontAddress": "0xf3fcd2c1a78f5eee",
				}))

		// test listing IDs separately, as they aren't stable
		assert.NotZero(t, scripts.ExtractUInt64ValueFromEvent(res,
			"A.01cf0e2f2f715450.SequelMarketplace.TokenListed", "listingID"))
		assert.NotZero(t, scripts.ExtractUInt64ValueFromEvent(res,
			"A.f8d6e0586b0a20c7.NFTStorefront.ListingAvailable", "listingResourceID"))
	})

	t.Run("Happy path (FUSD)", func(t *testing.T) {
		// Fund with Flow for FUSD setup fees
		scripts.FundAccountWithFlow(t, client, platformAcct.Address, "10.0")

		_ = se.NewTransaction("account_setup_fusd").SignProposeAndPayAs(sellerAcctName).Test(t).AssertSuccess()
		_ = se.NewTransaction("account_setup_fusd").SignProposeAndPayAs(platformAcct.Name).Test(t).AssertSuccess()

		res := se.NewTransaction("marketplace_list_fusd").
			SignProposeAndPayAs(sellerAcctName).
			UInt64Argument(nftID).
			UFix64Argument("200.0").
			Argument(cadence.NewOptional(cadence.String("link"))).
			Test(t).
			AssertSuccess().
			AssertPartialEvent(gwtf.NewTestEvent(
				"A.01cf0e2f2f715450.SequelMarketplace.TokenListed",
				map[string]interface{}{
					"asset":            "did:sequel:asset-id",
					"metadataLink":     "link",
					"nftID":            fmt.Sprintf("%d", nftID),
					"nftType":          "A.01cf0e2f2f715450.DigitalArt.NFT",
					"paymentVaultType": "A.f8d6e0586b0a20c7.FUSD.Vault",
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
							"receiver": "0xf3fcd2c1a78f5eee",
							"role":     "Owner",
						},
					},
					"price":             "200.00000000",
					"storefrontAddress": "0xf3fcd2c1a78f5eee",
				})).
			AssertPartialEvent(gwtf.NewTestEvent(
				"A.f8d6e0586b0a20c7.NFTStorefront.ListingAvailable",
				map[string]interface{}{
					"ftVaultType":       "Type\u003cA.f8d6e0586b0a20c7.FUSD.Vault\u003e()",
					"nftID":             fmt.Sprintf("%d", nftID),
					"nftType":           "Type\u003cA.01cf0e2f2f715450.DigitalArt.NFT\u003e()",
					"price":             "200.00000000",
					"storefrontAddress": "0xf3fcd2c1a78f5eee",
				}))

		// test listing IDs separately, as they aren't stable
		assert.NotZero(t, scripts.ExtractUInt64ValueFromEvent(res,
			"A.01cf0e2f2f715450.SequelMarketplace.TokenListed", "listingID"))
		assert.NotZero(t, scripts.ExtractUInt64ValueFromEvent(res,
			"A.f8d6e0586b0a20c7.NFTStorefront.ListingAvailable", "listingResourceID"))
	})
}

func TestMarketplace_buyToken(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true, true)
	require.NoError(t, err)

	scripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	platformAcct := client.Account(platformAccountName)

	// set up seller account

	sellerAcctName := "emulator-user1"
	sellerAcct := client.Account(sellerAcctName)

	scripts.FundAccountWithFlow(t, client, sellerAcct.Address, "10.0")

	_ = se.NewTransaction("account_setup").SignProposeAndPayAs(sellerAcctName).Test(t).AssertSuccess()

	// set up buyer account

	buyerAcctName := "emulator-user2"
	buyerAcct := client.Account(buyerAcctName)

	scripts.FundAccountWithFlow(t, client, buyerAcct.Address, "10.0")

	_ = se.NewTransaction("account_setup").SignProposeAndPayAs(buyerAcctName).Test(t).AssertSuccess()
	scripts.FundAccountWithFlow(t, client, buyerAcct.Address, "1000.0")

	metadata := SampleMetadata(1)
	profile := PrimaryOnlyEvergreenProfile(sellerAcct.Address, platformAcct.Address)

	_ = scripts.CreateSealDigitalArtTx(t, se, client, metadata, profile).
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

	nftID := scripts.ExtractUInt64ValueFromEvent(res,
		"A.01cf0e2f2f715450.DigitalArt.Minted", "id")

	// Assert that the account's collection is correct
	checkTokenInDigitalArtCollection(t, se, sellerAcct.Address.String(), nftID)
	checkDigitalArtCollectionLen(t, se, sellerAcct.Address.String(), 1)
	checkDigitalArtCollectionLen(t, se, buyerAcct.Address.String(), 0)

	res = se.NewTransaction("marketplace_list_flow").
		SignProposeAndPayAs(sellerAcctName).
		UInt64Argument(nftID).
		UFix64Argument("200.0").
		Argument(cadence.NewOptional(cadence.String("link"))).
		Test(t).
		AssertSuccess()

	listingID := scripts.ExtractUInt64ValueFromEvent(res,
		"A.f8d6e0586b0a20c7.NFTStorefront.ListingAvailable", "listingResourceID")

	t.Run("Happy path (Flow)", func(t *testing.T) {
		_ = se.NewTransaction("marketplace_buy_flow").
			SignProposeAndPayAs(buyerAcctName).
			UInt64Argument(listingID).
			Argument(cadence.NewAddress(sellerAcct.Address)).
			Argument(cadence.NewOptional(cadence.String("link"))).
			Test(t).
			AssertSuccess().
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.01cf0e2f2f715450.SequelMarketplace.TokenSold",
				map[string]interface{}{
					"listingID":         fmt.Sprintf("%d", listingID),
					"nftID":             fmt.Sprintf("%d", nftID),
					"nftType":           "A.01cf0e2f2f715450.DigitalArt.NFT",
					"paymentVaultType":  "A.0ae53cb6e3f42a79.FlowToken.Vault",
					"price":             "200.00000000",
					"storefrontAddress": "0xf3fcd2c1a78f5eee",
					"buyerAddress":      "0xe03daebed8ca0615",
					"metadataLink":      "link",
				}))

		// Assert that the account's collection is correct
		checkTokenInDigitalArtCollection(t, se, buyerAcct.Address.String(), 0)
		checkDigitalArtCollectionLen(t, se, buyerAcct.Address.String(), 1)
		checkDigitalArtCollectionLen(t, se, sellerAcct.Address.String(), 0)
	})
}

func TestMarketplace_payForMintedTokens(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true, true)
	require.NoError(t, err)

	scripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	scripts.PrepareFUSDMinter(t, se, client.Account("emulator-account").Address)

	evergreenAddr := flow.HexToAddress(se.WellKnownAddresses()["Evergreen"])

	buyerAcct := client.Account(user2AccountName)
	artistAcct := client.Account(user3AccountName) // the artist is the seller
	roleOneAcct := client.Account(user1AccountName)

	scripts.FundAccountWithFlow(t, client, buyerAcct.Address, "1000.0")

	happyPathProfile, err := evergreen.ProfileToCadence(&evergreen.Profile{
		ID: "did:sequel:evergreen3",
		Roles: []*evergreen.Role{
			{
				ID:                        "Artist",
				InitialSaleCommission:     0.8,
				SecondaryMarketCommission: 0.0,
				Address:                   artistAcct.Address,
			},
			{
				ID:                        "Role1",
				InitialSaleCommission:     0.2,
				SecondaryMarketCommission: 0.0,
				Address:                   roleOneAcct.Address,
			},
		},
	}, evergreenAddr)
	require.NoError(t, err)

	scriptWithFUSD := `
import FungibleToken from 0xee82856bf20e2aa6
import FUSD from 0xf8d6e0586b0a20c7
import Evergreen from 0x01cf0e2f2f715450
import SequelMarketplace from 0x01cf0e2f2f715450

transaction(numEditions: UInt64, unitPrice: UFix64, profile: Evergreen.Profile) {
    let paymentVault: @FungibleToken.Vault

    prepare(buyer: AuthAccount, platform: AuthAccount) {
        let mainVault = buyer.borrow<&FUSD.Vault>(from: /storage/fusdVault)
            ?? panic("Cannot borrow FUSD vault from acct storage")
        let price = unitPrice * UFix64(numEditions)
        self.paymentVault <- mainVault.withdraw(amount: price)
    }

    execute {
		SequelMarketplace.payForMintedTokens(
			unitPrice: unitPrice,
			numEditions: numEditions,
			sellerRole: "Artist",
			sellerVaultPath: /public/fusdReceiver,
			paymentVault: <-self.paymentVault,
			evergreenProfile: profile,
		)
   }
}`

	t.Run("Fail if seller's receiver is invalid", func(t *testing.T) {
		_ = se.NewTransaction("account_setup_fusd").SignProposeAndPayAs(buyerAcct.Name).Test(t).AssertSuccess()

		scripts.FundAccountWithFUSD(t, se, buyerAcct.Address, "1000.0")

		_ = client.Transaction(scriptWithFUSD).
			PayloadSigner(buyerAcct.Name).
			SignProposeAndPayAs(adminAccountName).
			UInt64Argument(1).
			UFix64Argument("100.0").
			Argument(happyPathProfile).
			Test(t).
			AssertFailure("missing fungible token receiver capability")
	})

	t.Run("If some receivers are invalid, send the remainder to last good receiver", func(t *testing.T) {
		// RoleOne's FUSD receiver is missing. RoleOne's cut will go to the seller (the artist).

		_ = se.NewTransaction("account_royalty_receiver_setup").SignAndProposeAs(artistAcct.Name).PayAs(adminAccountName).Test(t).AssertSuccess()

		_ = client.Transaction(scriptWithFUSD).
			PayloadSigner(buyerAcct.Name).
			SignProposeAndPayAs(adminAccountName).
			UInt64Argument(1).
			UFix64Argument("100.0").
			Argument(happyPathProfile).
			Test(t).
			AssertSuccess().
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.f8d6e0586b0a20c7.FUSD.TokensWithdrawn",
				map[string]interface{}{
					"amount": "100.00000000",
					"from":   "0x" + buyerAcct.Address.String(),
				})).
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.f8d6e0586b0a20c7.FUSD.TokensDeposited",
				map[string]interface{}{
					"amount": "80.00000000",
					"to":     "0x045a1763c93006ca",
				})).
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.f8d6e0586b0a20c7.FUSD.TokensDeposited",
				map[string]interface{}{
					"amount": "20.00000000",
					"to":     "0x045a1763c93006ca",
				}))
	})

	t.Run("Happy path (Flow)", func(t *testing.T) {
		_ = se.NewTransaction("account_royalty_receiver_setup").SignAndProposeAs(roleOneAcct.Name).PayAs(adminAccountName).Test(t).AssertSuccess()

		_ = client.Transaction(`
import FungibleToken from 0xee82856bf20e2aa6
import FlowToken from 0x0ae53cb6e3f42a79
import Evergreen from 0x01cf0e2f2f715450
import SequelMarketplace from 0x01cf0e2f2f715450

transaction(numEditions: UInt64, unitPrice: UFix64, profile: Evergreen.Profile) {
    let paymentVault: @FungibleToken.Vault

    prepare(buyer: AuthAccount, platform: AuthAccount) {
        let mainVault = buyer.borrow<&FlowToken.Vault>(from: /storage/flowTokenVault)
            ?? panic("Cannot borrow FlowToken vault from acct storage")
        let price = unitPrice * UFix64(numEditions)
        self.paymentVault <- mainVault.withdraw(amount: price)
    }

    execute {
		SequelMarketplace.payForMintedTokens(
			unitPrice: unitPrice,
			numEditions: numEditions,
			sellerRole: "Artist",
			sellerVaultPath: /public/flowTokenReceiver,
			paymentVault: <-self.paymentVault,
			evergreenProfile: profile,
		)
   }
}`).
			PayloadSigner(buyerAcct.Name).
			SignProposeAndPayAs(adminAccountName).
			UInt64Argument(1).
			UFix64Argument("100.0").
			Argument(happyPathProfile).
			Test(t).
			AssertSuccess().
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.0ae53cb6e3f42a79.FlowToken.TokensWithdrawn",
				map[string]interface{}{
					"amount": "100.00000000",
					"from":   "0x" + buyerAcct.Address.String(),
				})).
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.0ae53cb6e3f42a79.FlowToken.TokensDeposited",
				map[string]interface{}{
					"amount": "80.00000000",
					"to":     "0x045a1763c93006ca",
				})).
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.0ae53cb6e3f42a79.FlowToken.TokensDeposited",
				map[string]interface{}{
					"amount": "20.00000000",
					"to":     "0x" + roleOneAcct.Address.String(),
				}))
		require.NoError(t, err)
	})

	t.Run("Happy path (FUSD)", func(t *testing.T) {
		_ = client.Transaction(scriptWithFUSD).
			PayloadSigner(buyerAcct.Name).
			SignProposeAndPayAs(adminAccountName).
			UInt64Argument(1).
			UFix64Argument("100.0").
			Argument(happyPathProfile).
			Test(t).
			AssertSuccess().
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.f8d6e0586b0a20c7.FUSD.TokensWithdrawn",
				map[string]interface{}{
					"amount": "100.00000000",
					"from":   "0x" + buyerAcct.Address.String(),
				})).
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.f8d6e0586b0a20c7.FUSD.TokensDeposited",
				map[string]interface{}{
					"amount": "80.00000000",
					"to":     "0x045a1763c93006ca",
				})).
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.f8d6e0586b0a20c7.FUSD.TokensDeposited",
				map[string]interface{}{
					"amount": "20.00000000",
					"to":     "0x" + roleOneAcct.Address.String(),
				}))
	})
}

func TestMarketplace_withdrawToken(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true, true)
	require.NoError(t, err)

	scripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	platformAcct := client.Account(platformAccountName)

	// set up seller account

	sellerAcctName := "emulator-user1"
	sellerAcct := client.Account(sellerAcctName)

	scripts.FundAccountWithFlow(t, client, sellerAcct.Address, "10.0")

	_ = se.NewTransaction("account_setup").SignProposeAndPayAs(sellerAcctName).Test(t).AssertSuccess()

	metadata := SampleMetadata(1)
	profile := PrimaryOnlyEvergreenProfile(sellerAcct.Address, platformAcct.Address)

	_ = scripts.CreateSealDigitalArtTx(t, se, client, metadata, profile).
		SignProposeAndPayAs(adminAccountName).
		Test(t).
		AssertSuccess()

	_ = client.Transaction(se.GetStandardScript("digitalart_mint_edition")).
		SignProposeAndPayAs(adminAccountName).
		StringArgument(metadata.Asset).
		UInt64Argument(1).
		Argument(cadence.Address(sellerAcct.Address)).
		Test(t).
		AssertSuccess()

	var nftID uint64

	// Assert that the account's collection is correct
	checkTokenInDigitalArtCollection(t, se, sellerAcct.Address.String(), nftID)

	res := se.NewTransaction("marketplace_list_flow").
		SignProposeAndPayAs(sellerAcctName).
		UInt64Argument(nftID).
		UFix64Argument("200.0").
		Argument(cadence.NewOptional(cadence.String("link"))).
		Test(t).
		AssertSuccess()

	listingID := scripts.ExtractUInt64ValueFromEvent(res,
		"A.f8d6e0586b0a20c7.NFTStorefront.ListingAvailable", "listingResourceID")

	t.Run("Fail, if listing doesn't exist", func(t *testing.T) {
		_ = se.NewTransaction("marketplace_withdraw").
			SignProposeAndPayAs(sellerAcctName).
			UInt64Argument(12345).
			Test(t).
			AssertFailure("listing not found in Storefront")
	})

	t.Run("Happy path", func(t *testing.T) {
		_ = se.NewTransaction("marketplace_withdraw").
			SignProposeAndPayAs(sellerAcctName).
			UInt64Argument(listingID).
			Test(t).
			AssertSuccess().
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.01cf0e2f2f715450.SequelMarketplace.TokenWithdrawn",
				map[string]interface{}{
					"listingID":         fmt.Sprintf("%d", listingID),
					"nftID":             "0",
					"nftType":           "A.01cf0e2f2f715450.DigitalArt.NFT",
					"price":             "200.00000000",
					"storefrontAddress": "0xf3fcd2c1a78f5eee",
					"vaultType":         "A.0ae53cb6e3f42a79.FlowToken.Vault",
				}))

		// ensure the listing doesn't exist
		_, err = client.Script(`
import NFTStorefront from 0xf8d6e0586b0a20c7
import SequelMarketplace from 0x01cf0e2f2f715450

pub fun main(listingID:UInt64, storefrontAddress: Address) {
	let storefront = getAccount(storefrontAddress)
		.getCapability(NFTStorefront.StorefrontPublicPath)!
		.borrow<&NFTStorefront.Storefront{NFTStorefront.StorefrontPublic}>()
		?? panic("Could not borrow Storefront from provided address")

    if let listing = storefront.borrowListing(listingResourceID: listingID) {
		panic("listing still exists")
	}
}
`).
			UInt64Argument(listingID).
			Argument(cadence.Address(sellerAcct.Address)).
			RunReturns(context.Background())
		require.NoError(t, err)

		// ensure that the seller's collection still has the token
		checkTokenInDigitalArtCollection(t, se, sellerAcct.Address.String(), nftID)
	})
}

func TestMarketplace_buildPayments(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true, true)
	require.NoError(t, err)

	scripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	evergreenAddr := flow.HexToAddress(se.WellKnownAddresses()["Evergreen"])

	roleOneAcct := client.Account(user1AccountName)
	roleTwoAcct := client.Account(user2AccountName)
	sellerAcct := client.Account(user3AccountName)

	happyPathProfile, err := evergreen.ProfileToCadence(&evergreen.Profile{
		ID: "did:sequel:evergreen3",
		Roles: []*evergreen.Role{
			{
				ID:                        "Role1",
				InitialSaleCommission:     0.8,
				SecondaryMarketCommission: 0.05,
				Address:                   roleOneAcct.Address,
			},
			{
				ID:                        "Role2",
				InitialSaleCommission:     0.2,
				SecondaryMarketCommission: 0.025,
				Address:                   roleTwoAcct.Address,
			},
		},
	}, evergreenAddr)
	require.NoError(t, err)

	t.Run("Happy path (initial sale)", func(t *testing.T) {

		_, err = client.Script(`
import Evergreen from 0x01cf0e2f2f715450
import SequelMarketplace from 0x01cf0e2f2f715450

pub fun main(profile: Evergreen.Profile, seller: Address) {

	let instructions = SequelMarketplace.buildPayments(
        profile: profile,
        seller: seller,
        sellerRole: "Owner",
        sellerVaultPath: /public/flowTokenReceiver,
        price: 100.0,
        defaultReceiverPath: /public/flowTokenReceiver,
        initialSale: true,
        extraRoles: []
    )

	let payments = instructions.payments

	assert(payments != nil, message: "payments == nil")
	assert(payments.length == profile.roles.length, message: "incorrect number of payments")

	assert(payments[0].role == "Role1", message: "incorrect role 1")
	assert(payments[0].amount == 80.0, message: "incorrect amount 1")
	assert(payments[0].rate == 0.8, message: "incorrect rate 1")
	assert(payments[0].receiver == 0xf3fcd2c1a78f5eee, message: "incorrect receiver 1")

	assert(payments[1].role == "Role2", message: "incorrect role 2")
	assert(payments[1].amount == 20.0, message: "incorrect amount 2")
	assert(payments[1].rate == 0.2, message: "incorrect rate 2")
	assert(payments[1].receiver == 0xe03daebed8ca0615, message: "incorrect receiver 2")
}`).
			Argument(happyPathProfile).
			Argument(cadence.NewAddress(sellerAcct.Address)).
			RunReturns(context.Background())
		require.NoError(t, err)
	})

	t.Run("Happy path (secondary sale)", func(t *testing.T) {

		_, err = client.Script(`
import Evergreen from 0x01cf0e2f2f715450
import SequelMarketplace from 0x01cf0e2f2f715450

pub fun main(profile: Evergreen.Profile, seller: Address) {

	let instructions = SequelMarketplace.buildPayments(
        profile: profile,
        seller: seller,
        sellerRole: "Owner",
		sellerVaultPath: /public/flowTokenReceiver,
        price: 100.0,
		defaultReceiverPath: /public/flowTokenReceiver,
        initialSale: false,
        extraRoles: []
    )

	let payments = instructions.payments

	assert(payments != nil, message: "payments == nil")
	assert(payments.length == profile.roles.length+1, message: "incorrect number of payments")

	assert(payments[0].role == "Role1", message: "incorrect role 1")
	assert(payments[0].amount == 5.0, message: "incorrect amount 1")
	assert(payments[0].rate == 0.05, message: "incorrect rate 1")
	assert(payments[0].receiver == 0xf3fcd2c1a78f5eee, message: "incorrect receiver 1")

	assert(payments[1].role == "Role2", message: "incorrect role 2")
	assert(payments[1].amount == 2.5, message: "incorrect amount 2")
	assert(payments[1].rate == 0.025, message: "incorrect rate 2")
	assert(payments[1].receiver == 0xe03daebed8ca0615, message: "incorrect receiver 2")

	assert(payments[2].role == "Owner", message: "incorrect role 3")
	assert(payments[2].amount == 92.5, message: "incorrect amount 3")
	assert(payments[2].rate == 0.925, message: "incorrect rate 3")
	assert(payments[2].receiver == 0x045a1763c93006ca, message: "incorrect receiver 2")
}`).
			Argument(happyPathProfile).
			Argument(cadence.NewAddress(sellerAcct.Address)).
			RunReturns(context.Background())
		require.NoError(t, err)
	})

	t.Run("Should fail if any of the rates are out of range", func(t *testing.T) {
		profile := BasicEvergreenProfile(roleOneAcct.Address)
		profile.Roles[0].InitialSaleCommission = 1.25

		profileVal, err := evergreen.ProfileToCadence(profile, evergreenAddr)
		require.NoError(t, err)

		_, err = client.Script(`
import Evergreen from 0x01cf0e2f2f715450
import SequelMarketplace from 0x01cf0e2f2f715450

pub fun main(profile: Evergreen.Profile, seller: Address) {

	let instructions = SequelMarketplace.buildPayments(
        profile: profile,
        seller: seller,
        sellerRole: "Owner",
        price: 100.0,
		paymentVaultPath: /public/flowTokenReceiver,
        initialSale: true,
        extraRoles: []
    )
}`).
			Argument(profileVal).
			Argument(cadence.NewAddress(sellerAcct.Address)).
			RunReturns(context.Background())
		require.Error(t, err)
	})

	t.Run("Should fail if sum of rates is greater than 1.0", func(t *testing.T) {
		profile, err := evergreen.ProfileToCadence(&evergreen.Profile{
			ID: "did:sequel:evergreen3",
			Roles: []*evergreen.Role{
				{
					ID:                        "Role1",
					InitialSaleCommission:     0.8,
					SecondaryMarketCommission: 0.0,
					Address:                   roleOneAcct.Address,
				},
				{
					ID:                        "Role2",
					InitialSaleCommission:     0.8,
					SecondaryMarketCommission: 0.0,
					Address:                   roleTwoAcct.Address,
				},
			},
		}, evergreenAddr)
		require.NoError(t, err)

		_, err = client.Script(`
import Evergreen from 0x01cf0e2f2f715450
import SequelMarketplace from 0x01cf0e2f2f715450

pub fun main(profile: Evergreen.Profile, seller: Address) {

	let instructions = SequelMarketplace.buildPayments(
        profile: profile,
        seller: seller,
        sellerRole: "Owner",
        price: 100.0,
		paymentVaultPath: /public/flowTokenReceiver,
        initialSale: true,
        extraRoles: []
    )
}`).
			Argument(profile).
			Argument(cadence.NewAddress(sellerAcct.Address)).
			RunReturns(context.Background())
		require.Error(t, err)
	})

	t.Run("Roles with zero rate should not produce payments", func(t *testing.T) {
		profile, err := evergreen.ProfileToCadence(&evergreen.Profile{
			ID: "did:sequel:evergreen3",
			Roles: []*evergreen.Role{
				{
					ID:                        "Role1",
					InitialSaleCommission:     0.8,
					SecondaryMarketCommission: 0.05,
					Address:                   roleOneAcct.Address,
				},
				{
					ID:                        "Role2",
					InitialSaleCommission:     0.2,
					SecondaryMarketCommission: 0.0,
					Address:                   roleTwoAcct.Address,
				},
			},
		}, evergreenAddr)
		require.NoError(t, err)

		_, err = client.Script(`
import Evergreen from 0x01cf0e2f2f715450
import SequelMarketplace from 0x01cf0e2f2f715450

pub fun main(profile: Evergreen.Profile, seller: Address) {

	let instructions = SequelMarketplace.buildPayments(
        profile: profile,
        seller: seller,
        sellerRole: "Owner",
		sellerVaultPath: /public/flowTokenReceiver,
        price: 100.0,
		defaultReceiverPath: /public/flowTokenReceiver,
        initialSale: false,
        extraRoles: []
    )

	let payments = instructions.payments

	assert(payments != nil, message: "payments == nil")
	assert(payments.length == 2, message: "incorrect number of payments")

	assert(payments[0].role == "Role1", message: "incorrect role 1")
	assert(payments[0].amount == 5.0, message: "incorrect amount 1")
	assert(payments[0].rate == 0.05, message: "incorrect rate 1")
	assert(payments[0].receiver == 0xf3fcd2c1a78f5eee, message: "incorrect receiver 1")

	assert(payments[1].role == "Owner", message: "incorrect role 2")
	assert(payments[1].amount == 95.0, message: "incorrect amount 2")
	assert(payments[1].rate == 0.95, message: "incorrect rate 2")
	assert(payments[1].receiver == 0x045a1763c93006ca, message: "incorrect receiver 2")
}`).
			Argument(profile).
			Argument(cadence.NewAddress(sellerAcct.Address)).
			RunReturns(context.Background())
		require.NoError(t, err)
	})

	t.Run("Profile with no roles should allocate full amount to seller", func(t *testing.T) {
		profile, err := evergreen.ProfileToCadence(&evergreen.Profile{
			ID:    "did:sequel:evergreen3",
			Roles: []*evergreen.Role{},
		}, evergreenAddr)
		require.NoError(t, err)

		_, err = client.Script(`
import Evergreen from 0x01cf0e2f2f715450
import SequelMarketplace from 0x01cf0e2f2f715450

pub fun main(profile: Evergreen.Profile, seller: Address) {

	let instructions = SequelMarketplace.buildPayments(
        profile: profile,
        seller: seller,
        sellerRole: "Owner",
		sellerVaultPath: /public/flowTokenReceiver,
        price: 100.0,
		defaultReceiverPath: /public/flowTokenReceiver,
        initialSale: false,
        extraRoles: []
    )

	let payments = instructions.payments

	assert(payments != nil, message: "payments == nil")
	assert(payments.length == 1, message: "incorrect number of payments")

	assert(payments[0].role == "Owner", message: "incorrect role 1")
	assert(payments[0].amount == 100.0, message: "incorrect amount 1")
	assert(payments[0].rate == 1.0, message: "incorrect rate 1")
	assert(payments[0].receiver == 0x045a1763c93006ca, message: "incorrect receiver 1")
}`).
			Argument(profile).
			Argument(cadence.NewAddress(sellerAcct.Address)).
			RunReturns(context.Background())
		require.NoError(t, err)
	})

	t.Run("Extra roles should produce additional payments", func(t *testing.T) {
		extraTwoAcct := client.Account(platformAccountName)

		profile, err := evergreen.ProfileToCadence(&evergreen.Profile{
			ID: "did:sequel:evergreen3",
			Roles: []*evergreen.Role{
				{
					ID:                        "Role1",
					InitialSaleCommission:     1.0,
					SecondaryMarketCommission: 0.05,
					Address:                   roleOneAcct.Address,
				},
			},
		}, evergreenAddr)
		require.NoError(t, err)

		_, err = client.Script(`
import Evergreen from 0x01cf0e2f2f715450
import SequelMarketplace from 0x01cf0e2f2f715450

pub fun main(profile: Evergreen.Profile, seller: Address, extra1: Address, extra2: Address) {

	let instructions = SequelMarketplace.buildPayments(
        profile: profile,
        seller: seller,
        sellerRole: "Owner",
		sellerVaultPath: /public/flowTokenReceiver,
        price: 100.0,
		defaultReceiverPath: /public/flowTokenReceiver,
        initialSale: false,
        extraRoles: [
			Evergreen.Role(
				id: "Extra1",
				description: "Extra 1",
				initialSaleCommission: UFix64(0.0),
				secondaryMarketCommission: UFix64(0.02),
				address: extra1,
				receiverPath: nil
			),
			Evergreen.Role(
				id: "Extra2",
				description: "Extra 2",
				initialSaleCommission: UFix64(0.0),
				secondaryMarketCommission: UFix64(0.04),
				address: extra2,
				receiverPath: nil
			)
		]
    )

	let payments = instructions.payments

	assert(payments != nil, message: "payments == nil")
	assert(payments.length == 4, message: "incorrect number of payments")

	assert(payments[0].role == "Role1", message: "incorrect role 1")
	assert(payments[0].amount == 5.0, message: "incorrect amount 1")
	assert(payments[0].rate == 0.05, message: "incorrect rate 1")
	assert(payments[0].receiver == 0xf3fcd2c1a78f5eee, message: "incorrect receiver 1")

	assert(payments[1].role == "Extra1", message: "incorrect role 2")
	assert(payments[1].amount == 2.0, message: "incorrect amount 2")
	assert(payments[1].rate == 0.02, message: "incorrect rate 2")
	assert(payments[1].receiver == 0xe03daebed8ca0615, message: "incorrect receiver 2")

	assert(payments[2].role == "Extra2", message: "incorrect role 3")
	assert(payments[2].amount == 4.0, message: "incorrect amount 3")
	assert(payments[2].rate == 0.04, message: "incorrect rate 3")
	assert(payments[2].receiver == 0x179b6b1cb6755e31, message: "incorrect receiver 3")

	assert(payments[3].role == "Owner", message: "incorrect role 4")
	assert(payments[3].amount == 89.0, message: "incorrect amount 4")
	assert(payments[3].rate == 0.89, message: "incorrect rate 4")
	assert(payments[3].receiver == 0x045a1763c93006ca, message: "incorrect receiver 4")
}`).
			Argument(profile).
			Argument(cadence.NewAddress(sellerAcct.Address)).
			Argument(cadence.NewAddress(roleTwoAcct.Address)).
			Argument(cadence.NewAddress(extraTwoAcct.Address)).
			RunReturns(context.Background())
		require.NoError(t, err)
	})
}
