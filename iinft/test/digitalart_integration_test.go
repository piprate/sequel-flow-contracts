package test

import (
	"context"
	"testing"

	"github.com/onflow/cadence"
	"github.com/piprate/sequel-flow-contracts/iinft"
	"github.com/piprate/sequel-flow-contracts/iinft/evergreen"
	"github.com/piprate/sequel-flow-contracts/iinft/testscripts"
	"github.com/piprate/splash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDigitalArt_Integration_MintOnDemand_ExampleToken(t *testing.T) {
	client, err := splash.NewInMemoryTestConnector("../..", true)
	require.NoError(t, err)

	testscripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := iinft.NewTemplateEngine(client)
	require.NoError(t, err)

	// set up platform account

	platformAcct := client.Account(platformAccountName)

	testscripts.SetUpRoyaltyReceivers(t, se, platformAccountName, adminAccountName, "ExampleToken")

	// set up green account

	greenAcctName := user3AccountName
	greenAcct := client.Account(greenAcctName)

	testscripts.SetUpRoyaltyReceivers(t, se, greenAcctName, adminAccountName, "ExampleToken")

	// set up artist account

	artistAcctName := user1AccountName
	artistAcct := client.Account(artistAcctName)

	testscripts.SetUpRoyaltyReceivers(t, se, artistAcctName, adminAccountName, "ExampleToken")

	// set up buyer account

	buyerAcctName := user2AccountName
	buyerAcct := client.Account(buyerAcctName)

	testscripts.FundAccountWithFlow(t, se, buyerAcct.Address, "10.0")

	_ = se.NewTransaction("account_setup").SignProposeAndPayAs(buyerAcctName).Test(t).AssertSuccess()
	_ = se.NewTransaction("account_setup_example_ft").SignProposeAndPayAs(buyerAcctName).Test(t).AssertSuccess()
	testscripts.FundAccountWithExampleToken(t, se, buyerAcct.Address, "1000.0")

	checkDigitalArtNFTSupply(t, se, 0)
	checkDigitalArtCollectionLen(t, se, buyerAcct.Address.String(), 0)

	metadata := SampleMetadata(4)
	profile := &evergreen.Profile{
		ID: "did:sequel:evergreen1",
		Roles: []*evergreen.Role{
			{
				ID:                        evergreen.RoleArtist,
				InitialSaleCommission:     0.9,
				SecondaryMarketCommission: 0.025,
				Address:                   artistAcct.Address,
			},
			{
				ID:                        evergreen.RolePlatform,
				InitialSaleCommission:     0.05,
				SecondaryMarketCommission: 0.025,
				Address:                   platformAcct.Address,
			},
			{
				ID:                        "ClimateActionFund",
				InitialSaleCommission:     0.05,
				SecondaryMarketCommission: 0.025,
				Address:                   greenAcct.Address,
			},
		},
	}

	t.Run("Should be able to mint a token on demand (master not sealed)", func(t *testing.T) {

		_ = client.Transaction(se.GetCustomScript("digitalart_mint_on_demand", iinft.MintOnDemandParameters{
			Metadata: metadata,
			Profile:  profile,
		})).
			PayloadSigner(buyerAcctName).
			SignProposeAndPayAs(adminAccountName).
			StringArgument(metadata.Asset).
			UInt64Argument(1).
			UFix64Argument("100.0").
			Argument(cadence.NewAddress(se.ContractAddress("ExampleToken"))).
			StringArgument("ExampleToken").
			UInt64Argument(123).
			Test(t).
			AssertSuccess().
			AssertEventCount(15).
			AssertEmitEventName(
				"A.179b6b1cb6755e31.DigitalArt.Minted",
				"A.179b6b1cb6755e31.DigitalArt.Deposit",
				"A.ee82856bf20e2aa6.FungibleToken.Withdrawn",
				"A.ee82856bf20e2aa6.FungibleToken.Deposited").
			AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.DigitalArt.Minted", map[string]interface{}{
				"id":      "0",
				"asset":   "did:sequel:asset-id",
				"edition": "1",
				"modID":   "123",
			})).
			AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.DigitalArt.Deposit", map[string]interface{}{
				"id": "0",
				"to": "0x045a1763c93006ca",
			}))

		// Assert that the account's collection is correct
		checkTokenInDigitalArtCollection(t, se, buyerAcct.Address.String(), 0)
		checkDigitalArtCollectionLen(t, se, buyerAcct.Address.String(), 1)
		checkDigitalArtNFTSupply(t, se, 1)

		val, err := se.NewScript("digitalart_get_metadata").
			Argument(cadence.NewAddress(buyerAcct.Address)).
			UInt64Argument(0).
			RunReturns(context.Background())
		require.NoError(t, err)

		meta, err := iinft.DigitalArtMetadataFromCadence(val)
		require.NoError(t, err)

		assert.Equal(t, uint64(1), meta.Edition)

		assert.Equal(t, 90.0, testscripts.GetExampleTokenBalance(t, se, artistAcct.Address))
		assert.Equal(t, 900.0, testscripts.GetExampleTokenBalance(t, se, buyerAcct.Address))
		assert.Equal(t, 5.0, testscripts.GetExampleTokenBalance(t, se, platformAcct.Address))
		assert.Equal(t, 5.0, testscripts.GetExampleTokenBalance(t, se, greenAcct.Address))
	})

	t.Run("Should be able to mint a token on demand (master sealed, metadata provided)", func(t *testing.T) {

		_ = client.Transaction(se.GetCustomScript("digitalart_mint_on_demand", iinft.MintOnDemandParameters{
			Metadata: metadata,
			Profile:  profile,
		})).
			PayloadSigner(buyerAcctName).
			SignProposeAndPayAs(adminAccountName).
			StringArgument(metadata.Asset).
			UInt64Argument(1).
			UFix64Argument("100.0").
			Argument(cadence.NewAddress(se.ContractAddress("ExampleToken"))).
			StringArgument("ExampleToken").
			UInt64Argument(123).
			Test(t).
			AssertSuccess().
			AssertEventCount(15).
			AssertEmitEventName(
				"A.179b6b1cb6755e31.DigitalArt.Minted",
				"A.179b6b1cb6755e31.DigitalArt.Deposit",
				"A.ee82856bf20e2aa6.FungibleToken.Withdrawn",
				"A.ee82856bf20e2aa6.FungibleToken.Deposited").
			AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.DigitalArt.Minted", map[string]interface{}{
				"id":      "1",
				"asset":   "did:sequel:asset-id",
				"edition": "2",
				"modID":   "123",
			})).
			AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.DigitalArt.Deposit", map[string]interface{}{
				"id": "1",
				"to": "0x045a1763c93006ca",
			}))

		// Assert that the account's collection is correct
		checkTokenInDigitalArtCollection(t, se, buyerAcct.Address.String(), 1)
		checkDigitalArtCollectionLen(t, se, buyerAcct.Address.String(), 2)
		checkDigitalArtNFTSupply(t, se, 2)

		val, err := se.NewScript("digitalart_get_metadata").
			Argument(cadence.NewAddress(buyerAcct.Address)).
			UInt64Argument(0).
			RunReturns(context.Background())
		require.NoError(t, err)

		meta, err := iinft.DigitalArtMetadataFromCadence(val)
		require.NoError(t, err)

		assert.Equal(t, uint64(1), meta.Edition)

		assert.Equal(t, 180.0, testscripts.GetExampleTokenBalance(t, se, artistAcct.Address))
		assert.Equal(t, 800.0, testscripts.GetExampleTokenBalance(t, se, buyerAcct.Address))
		assert.Equal(t, 10.0, testscripts.GetExampleTokenBalance(t, se, platformAcct.Address))
		assert.Equal(t, 10.0, testscripts.GetExampleTokenBalance(t, se, greenAcct.Address))
	})

	t.Run("Should be able to mint a token on demand (master sealed, metadata not provided)", func(t *testing.T) {

		_ = client.Transaction(se.GetCustomScript("digitalart_mint_on_demand", iinft.MintOnDemandParameters{})).
			PayloadSigner(buyerAcctName).
			SignProposeAndPayAs(adminAccountName).
			StringArgument(metadata.Asset).
			UInt64Argument(1).
			UFix64Argument("100.0").
			Argument(cadence.NewAddress(se.ContractAddress("ExampleToken"))).
			StringArgument("ExampleToken").
			UInt64Argument(123).
			Test(t).
			AssertSuccess().
			AssertEventCount(15).
			AssertEmitEventName(
				"A.179b6b1cb6755e31.DigitalArt.Minted",
				"A.179b6b1cb6755e31.DigitalArt.Deposit",
				"A.ee82856bf20e2aa6.FungibleToken.Withdrawn",
				"A.ee82856bf20e2aa6.FungibleToken.Deposited").
			AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.DigitalArt.Minted", map[string]interface{}{
				"asset":   "did:sequel:asset-id",
				"edition": "3",
				"id":      "2",
				"modID":   "123",
			})).
			AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.DigitalArt.Deposit", map[string]interface{}{
				"id": "2",
				"to": "0x045a1763c93006ca",
			}))

		// Assert that the account's collection is correct
		checkTokenInDigitalArtCollection(t, se, buyerAcct.Address.String(), 2)
		checkDigitalArtCollectionLen(t, se, buyerAcct.Address.String(), 3)
		checkDigitalArtNFTSupply(t, se, 3)

		val, err := se.NewScript("digitalart_get_metadata").
			Argument(cadence.NewAddress(buyerAcct.Address)).
			UInt64Argument(0).
			RunReturns(context.Background())
		require.NoError(t, err)

		meta, err := iinft.DigitalArtMetadataFromCadence(val)
		require.NoError(t, err)

		assert.Equal(t, uint64(1), meta.Edition)

		assert.Equal(t, 270.0, testscripts.GetExampleTokenBalance(t, se, artistAcct.Address))
		assert.Equal(t, 700.0, testscripts.GetExampleTokenBalance(t, se, buyerAcct.Address))
		assert.Equal(t, 15.0, testscripts.GetExampleTokenBalance(t, se, platformAcct.Address))
		assert.Equal(t, 15.0, testscripts.GetExampleTokenBalance(t, se, greenAcct.Address))
	})
}

func TestDigitalArt_Integration_MintOnDemand_Flow(t *testing.T) {
	client, err := splash.NewInMemoryTestConnector("../..", true)
	require.NoError(t, err)

	testscripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := iinft.NewTemplateEngine(client)
	require.NoError(t, err)

	// set up platform account

	platformAcct := client.Account(platformAccountName)

	testscripts.SetUpRoyaltyReceivers(t, se, platformAccountName, adminAccountName)

	// set up green account

	greenAcctName := user3AccountName
	greenAcct := client.Account(greenAcctName)

	testscripts.SetUpRoyaltyReceivers(t, se, greenAcctName, adminAccountName)

	// set up artist account

	artistAcctName := user1AccountName
	artistAcct := client.Account(artistAcctName)

	testscripts.SetUpRoyaltyReceivers(t, se, artistAcctName, adminAccountName)

	// set up buyer account

	buyerAcctName := user2AccountName
	buyerAcct := client.Account(buyerAcctName)

	testscripts.FundAccountWithFlow(t, se, buyerAcct.Address, "1000.0")

	_ = se.NewTransaction("account_setup").SignProposeAndPayAs(buyerAcctName).Test(t).AssertSuccess()

	checkDigitalArtNFTSupply(t, se, 0)
	checkDigitalArtCollectionLen(t, se, buyerAcct.Address.String(), 0)

	metadata := SampleMetadata(4)
	profile := &evergreen.Profile{
		ID: "did:sequel:evergreen1",
		Roles: []*evergreen.Role{
			{
				ID:                        evergreen.RoleArtist,
				InitialSaleCommission:     0.9,
				SecondaryMarketCommission: 0.025,
				Address:                   artistAcct.Address,
			},
			{
				ID:                        evergreen.RolePlatform,
				InitialSaleCommission:     0.05,
				SecondaryMarketCommission: 0.025,
				Address:                   platformAcct.Address,
			},
			{
				ID:                        "ClimateActionFund",
				InitialSaleCommission:     0.05,
				SecondaryMarketCommission: 0.025,
				Address:                   greenAcct.Address,
			},
		},
	}

	t.Run("Should be able to mint a token on demand (master not sealed)", func(t *testing.T) {

		_ = client.Transaction(se.GetCustomScript("digitalart_mint_on_demand", iinft.MintOnDemandParameters{
			Metadata: metadata,
			Profile:  profile,
		})).
			PayloadSigner(buyerAcctName).
			SignProposeAndPayAs(adminAccountName).
			StringArgument(metadata.Asset).
			UInt64Argument(1).
			UFix64Argument("100.0").
			Argument(cadence.NewAddress(se.ContractAddress("FlowToken"))).
			StringArgument("FlowToken").
			UInt64Argument(123).
			Test(t).
			AssertSuccess().
			AssertEventCount(22).
			AssertEmitEventName(
				"A.179b6b1cb6755e31.DigitalArt.Minted",
				"A.179b6b1cb6755e31.DigitalArt.Deposit",
				"A.0ae53cb6e3f42a79.FlowToken.TokensWithdrawn",
				"A.0ae53cb6e3f42a79.FlowToken.TokensDeposited").
			AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.DigitalArt.Minted", map[string]interface{}{
				"id":      "0",
				"asset":   "did:sequel:asset-id",
				"edition": "1",
				"modID":   "123",
			})).
			AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.DigitalArt.Deposit", map[string]interface{}{
				"id": "0",
				"to": "0x045a1763c93006ca",
			}))

		// Assert that the account's collection is correct
		checkTokenInDigitalArtCollection(t, se, buyerAcct.Address.String(), 0)
		checkDigitalArtCollectionLen(t, se, buyerAcct.Address.String(), 1)
		checkDigitalArtNFTSupply(t, se, 1)

		val, err := se.NewScript("digitalart_get_metadata").
			Argument(cadence.NewAddress(buyerAcct.Address)).
			UInt64Argument(0).
			RunReturns(context.Background())
		require.NoError(t, err)

		meta, err := iinft.DigitalArtMetadataFromCadence(val)
		require.NoError(t, err)

		assert.Equal(t, uint64(1), meta.Edition)

		assert.InDelta(t, initialFlowBalance+1000.0-100.0, testscripts.GetFlowBalance(t, se, buyerAcct.Address), 0.001)
		assert.Equal(t, initialFlowBalance+90.0, testscripts.GetFlowBalance(t, se, artistAcct.Address))
		assert.Equal(t, initialFlowBalance+5.0, testscripts.GetFlowBalance(t, se, platformAcct.Address))
		assert.Equal(t, initialFlowBalance+5.0, testscripts.GetFlowBalance(t, se, greenAcct.Address))
	})

	t.Run("Should be able to mint a token on demand (master sealed)", func(t *testing.T) {

		_ = client.Transaction(se.GetCustomScript("digitalart_mint_on_demand", iinft.MintOnDemandParameters{
			Metadata: metadata,
			Profile:  profile,
		})).
			PayloadSigner(buyerAcctName).
			SignProposeAndPayAs(adminAccountName).
			StringArgument(metadata.Asset).
			UInt64Argument(1).
			UFix64Argument("100.0").
			Argument(cadence.NewAddress(se.ContractAddress("FlowToken"))).
			StringArgument("FlowToken").
			UInt64Argument(123).
			Test(t).
			AssertSuccess().
			AssertEventCount(22).
			AssertEmitEventName(
				"A.179b6b1cb6755e31.DigitalArt.Minted",
				"A.179b6b1cb6755e31.DigitalArt.Deposit",
				"A.0ae53cb6e3f42a79.FlowToken.TokensWithdrawn",
				"A.0ae53cb6e3f42a79.FlowToken.TokensDeposited").
			AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.DigitalArt.Minted", map[string]interface{}{
				"id":      "1",
				"asset":   "did:sequel:asset-id",
				"edition": "2",
				"modID":   "123",
			})).
			AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.DigitalArt.Deposit", map[string]interface{}{
				"id": "1",
				"to": "0x045a1763c93006ca",
			}))

		// Assert that the account's collection is correct
		checkTokenInDigitalArtCollection(t, se, buyerAcct.Address.String(), 1)
		checkDigitalArtCollectionLen(t, se, buyerAcct.Address.String(), 2)
		checkDigitalArtNFTSupply(t, se, 2)

		val, err := se.NewScript("digitalart_get_metadata").
			Argument(cadence.NewAddress(buyerAcct.Address)).
			UInt64Argument(0).
			RunReturns(context.Background())
		require.NoError(t, err)

		meta, err := iinft.DigitalArtMetadataFromCadence(val)
		require.NoError(t, err)

		assert.Equal(t, uint64(1), meta.Edition)

		assert.InDelta(t, initialFlowBalance+1000.0-2*100.0, testscripts.GetFlowBalance(t, se, buyerAcct.Address), 0.001)
		assert.Equal(t, initialFlowBalance+2*90.0, testscripts.GetFlowBalance(t, se, artistAcct.Address))
		assert.Equal(t, initialFlowBalance+2*5.0, testscripts.GetFlowBalance(t, se, platformAcct.Address))
		assert.Equal(t, initialFlowBalance+2*5.0, testscripts.GetFlowBalance(t, se, greenAcct.Address))
	})
}

func TestDigitalArt_Integration_Transfer(t *testing.T) {
	// this test executes:
	//   * 'withdraw' and 'deposit' methods of DigitalArt.Collection
	//   * the script (digitalart_destroy) for destroying a token

	client, err := splash.NewInMemoryTestConnector("../..", true)
	require.NoError(t, err)

	testscripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := iinft.NewTemplateEngine(client)
	require.NoError(t, err)

	senderAcctName := user1AccountName
	senderAcct := client.Account(senderAcctName)

	testscripts.FundAccountWithFlow(t, se, senderAcct.Address, "10.0")

	_ = se.NewTransaction("account_setup").
		SignProposeAndPayAs(senderAcctName).
		Test(t).
		AssertSuccess()

	receiverAcctName := user2AccountName
	receiverAcct := client.Account(receiverAcctName)

	testscripts.FundAccountWithFlow(t, se, receiverAcct.Address, "10.0")

	_ = se.NewTransaction("account_setup").
		SignProposeAndPayAs(receiverAcctName).
		Test(t).
		AssertSuccess()

	metadata := SampleMetadata(4)
	profile := BasicEvergreenProfile(senderAcct.Address)

	_ = testscripts.CreateSealDigitalArtTx(t, se, client, metadata, profile).
		SignProposeAndPayAs(adminAccountName).
		Test(t).
		AssertSuccess()

	_ = client.Transaction(se.GetStandardScript("digitalart_mint_edition")).
		SignProposeAndPayAs(adminAccountName).
		StringArgument(metadata.Asset).
		UInt64Argument(1).
		Argument(cadence.Address(senderAcct.Address)).
		Test(t).
		AssertSuccess()

	t.Run("Shouldn't be able to withdraw an NFT that doesn't exist in a collection", func(t *testing.T) {

		_ = se.NewTransaction("digitalart_transfer").
			SignProposeAndPayAs(senderAcctName).
			UInt64Argument(3).
			Argument(cadence.Address(receiverAcct.Address)).
			Test(t).
			AssertFailure("Could not withdraw an NFT with ID")

		checkDigitalArtCollectionLen(t, se, receiverAcct.Address.String(), 0)
		checkDigitalArtCollectionLen(t, se, senderAcct.Address.String(), 1)
	})

	t.Run("Should be able to withdraw an NFT and deposit to another accounts collection", func(t *testing.T) {
		_ = se.NewTransaction("digitalart_transfer").
			SignProposeAndPayAs(senderAcctName).
			UInt64Argument(0).
			Argument(cadence.Address(receiverAcct.Address)).
			Test(t).
			AssertSuccess()

		// Assert that the account's collection is correct
		checkTokenInDigitalArtCollection(t, se, receiverAcct.Address.String(), 0)
		checkDigitalArtCollectionLen(t, se, receiverAcct.Address.String(), 1)
		checkDigitalArtCollectionLen(t, se, senderAcct.Address.String(), 0)
	})

	t.Run("Should be able to withdraw an NFT and destroy it, not reducing the supply", func(t *testing.T) {

		_ = se.NewTransaction("digitalart_destroy").
			SignProposeAndPayAs(receiverAcctName).
			UInt64Argument(0).
			Test(t).
			AssertSuccess()

		checkDigitalArtCollectionLen(t, se, receiverAcct.Address.String(), 0)
		checkDigitalArtCollectionLen(t, se, senderAcct.Address.String(), 0)
		checkDigitalArtNFTSupply(t, se, 1)
	})
}
