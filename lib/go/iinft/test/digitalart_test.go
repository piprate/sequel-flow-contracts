package test

import (
	"os"
	"testing"
	"time"

	"github.com/onflow/cadence"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/evergreen"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/gwtf"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/scripts"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	adminAccount = "emulator-sequel-admin"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp})
}

func TestSealDigitalArtMaster(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true, true)
	require.NoError(t, err)

	scripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	userAcct := client.Account("emulator-user1")

	sampleMetadata := &iinft.DigitalArtMetadata{
		MetadataURI:       "ipfs://QmMetadata",
		Name:              "Pure Art",
		Artist:            "did:sequel:artist",
		Description:       "Digital art in its purest form",
		Type:              "Image",
		ContentURI:        "ipfs://QmContent",
		ContentPreviewURI: "ipfs://QmPreview",
		ContentMimetype:   "image/jpeg",
		MaxEdition:        4,
		Asset:             "did:sequel:asset-id",
		Record:            "record-id",
		AssetHead:         "asset-head-id",
	}

	profile := &evergreen.Profile{
		ID: 0,
		Roles: []*evergreen.Role{
			{
				Role:                      evergreen.RoleArtist,
				InitialSaleCommission:     0.8,
				SecondaryMarketCommission: 0.2,
				Address:                   userAcct.Address(),
			},
		},
	}

	t.Run("Should be able to seal new digital art master", func(t *testing.T) {

		_ = scripts.CreateSealDigitalArtTx(se, client, sampleMetadata, profile).
			SignProposeAndPayAs(adminAccount).
			Test(t).
			AssertSuccess()
	})

	t.Run("Shouldn't be able to seal the same digital art master twice", func(t *testing.T) {

		sampleMetadata2 := *sampleMetadata
		sampleMetadata2.Asset = "did:sequel:asset-2"

		// Seal the master

		_ = scripts.CreateSealDigitalArtTx(se, client, &sampleMetadata2, profile).
			SignProposeAndPayAs(adminAccount).
			Test(t).
			AssertSuccess()

		// try again
		_ = scripts.CreateSealDigitalArtTx(se, client, &sampleMetadata2, profile).
			SignProposeAndPayAs(adminAccount).
			Test(t).
			AssertFailure("master already sealed")
	})
}

func TestMintDigitalArtEditions(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true, true)
	require.NoError(t, err)

	scripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	userAcct := client.Account("emulator-user1")

	scripts.FundAccountWithFlow(t, client, userAcct.Address(), "10.0")

	_ = se.NewTransaction("account_setup").
		SignProposeAndPayAs("emulator-user1").
		Test(t).
		AssertSuccess()

	checkDigitalArtNFTSupply(t, se, 0)
	checkDigitalArtCollectionLen(t, se, userAcct.Address().String(), 0)

	metadata := &iinft.DigitalArtMetadata{
		MetadataURI:       "ipfs://QmMetadata",
		Name:              "Pure Art",
		Artist:            "did:sequel:artist",
		Description:       "Digital art in its purest form",
		Type:              "Image",
		ContentURI:        "ipfs://QmContent",
		ContentPreviewURI: "ipfs://QmPreview",
		ContentMimetype:   "image/jpeg",
		MaxEdition:        4,
		Asset:             "did:sequel:asset-id",
		Record:            "record-id",
		AssetHead:         "asset-head-id",
	}

	profile := &evergreen.Profile{
		ID: 0,
		Roles: []*evergreen.Role{
			{
				Role:                      evergreen.RoleArtist,
				InitialSaleCommission:     0.8,
				SecondaryMarketCommission: 0.2,
				Address:                   userAcct.Address(),
			},
		},
	}

	_ = scripts.CreateSealDigitalArtTx(se, client, metadata, profile).
		SignProposeAndPayAs(adminAccount).
		Test(t).
		AssertSuccess()

	t.Run("Should be able to mint a token", func(t *testing.T) {

		_ = client.Transaction(se.GetStandardScript("digitalart_mint_edition")).
			SignProposeAndPayAs(adminAccount).
			StringArgument(metadata.Asset).
			UInt64Argument(1).
			Argument(cadence.Address(userAcct.Address())).
			Test(t).
			AssertSuccess().
			AssertEventCount(5).
			AssertEmitEventName("A.01cf0e2f2f715450.DigitalArt.Minted", "A.01cf0e2f2f715450.DigitalArt.Deposit").
			AssertEmitEvent(gwtf.NewTestEvent("A.01cf0e2f2f715450.DigitalArt.Minted", map[string]interface{}{
				"id":      "0",
				"asset":   "did:sequel:asset-id",
				"edition": "1",
				"modID":   "0",
			})).
			AssertEmitEvent(gwtf.NewTestEvent("A.01cf0e2f2f715450.DigitalArt.Deposit", map[string]interface{}{
				"id": "0",
				"to": "0xf3fcd2c1a78f5eee",
			}))

		// Assert that the account's collection is correct
		checkTokenInDigitalArtCollection(t, se, userAcct.Address().String(), 0)
		checkDigitalArtCollectionLen(t, se, userAcct.Address().String(), 1)
		checkDigitalArtNFTSupply(t, se, 1)

		val, err := se.NewScript("digitalart_get_metadata").
			Argument(cadence.NewAddress(userAcct.Address())).
			UInt64Argument(0).
			RunReturns()
		require.NoError(t, err)

		meta, err := iinft.DigitalArtMetadataFromCadence(val)
		require.NoError(t, err)

		assert.Equal(t, uint64(1), meta.Edition)
	})

	t.Run("Editions should have different metadata", func(t *testing.T) {
		_ = client.Transaction(se.GetStandardScript("digitalart_mint_edition")).
			SignProposeAndPayAs(adminAccount).
			StringArgument(metadata.Asset).
			UInt64Argument(1).
			Argument(cadence.Address(userAcct.Address())).
			Test(t).
			AssertSuccess().
			AssertEventCount(5).
			AssertEmitEventName("A.01cf0e2f2f715450.DigitalArt.Minted", "A.01cf0e2f2f715450.DigitalArt.Deposit").
			AssertEmitEvent(gwtf.NewTestEvent("A.01cf0e2f2f715450.DigitalArt.Minted", map[string]interface{}{
				"id":      "1",
				"asset":   "did:sequel:asset-id",
				"edition": "2",
				"modID":   "0",
			})).
			AssertEmitEvent(gwtf.NewTestEvent("A.01cf0e2f2f715450.DigitalArt.Deposit", map[string]interface{}{
				"id": "1",
				"to": "0xf3fcd2c1a78f5eee",
			}))

		// Assert that the account's collection is correct
		checkTokenInDigitalArtCollection(t, se, userAcct.Address().String(), 1)
		checkDigitalArtCollectionLen(t, se, userAcct.Address().String(), 2)
		checkDigitalArtNFTSupply(t, se, 2)

		val, err := se.NewScript("digitalart_get_metadata").
			Argument(cadence.NewAddress(userAcct.Address())).
			UInt64Argument(1).
			RunReturns()
		require.NoError(t, err)

		meta, err := iinft.DigitalArtMetadataFromCadence(val)
		require.NoError(t, err)

		assert.Equal(t, uint64(2), meta.Edition)
	})

	t.Run("Shouldn't be able to borrow a reference to an NFT that doesn't exist", func(t *testing.T) {

		// test for non-existent token
		_, err := se.NewInlineScript(
			inspectCollectionScript(se.WellKnownAddresses(), userAcct.Address().String(),
				"DigitalArt", "DigitalArt.CollectionPublicPath", 5),
		).RunReturns()
		require.Error(t, err)
	})
}

func TestMintDigitalArtEditionsOnDemandFUSD(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true, true)
	require.NoError(t, err)

	scripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	scripts.PrepareFUSDMinter(t, se, client.Account("emulator-account").Address())

	// set up platform account

	platformAcctName := "emulator-sequel-platform"
	platformAcct := client.Account(platformAcctName)

	scripts.FundAccountWithFlow(t, client, platformAcct.Address(), "10.0")

	_ = se.NewTransaction("account_setup_fusd").SignProposeAndPayAs(platformAcctName).Test(t).AssertSuccess()

	// set up green account

	greenAcctName := "emulator-user3"
	greenAcct := client.Account(greenAcctName)

	scripts.FundAccountWithFlow(t, client, greenAcct.Address(), "10.0")

	_ = se.NewTransaction("account_setup_fusd").SignProposeAndPayAs(greenAcctName).Test(t).AssertSuccess()

	// set up artist account

	artistAcctName := "emulator-user1"
	artistAcct := client.Account(artistAcctName)

	scripts.FundAccountWithFlow(t, client, artistAcct.Address(), "10.0")

	_ = se.NewTransaction("account_setup_fusd").SignProposeAndPayAs(artistAcctName).Test(t).AssertSuccess()

	// set up buyer account

	buyerAcctName := "emulator-user2"
	buyerAcct := client.Account(buyerAcctName)

	scripts.FundAccountWithFlow(t, client, buyerAcct.Address(), "10.0")

	_ = se.NewTransaction("account_setup").SignProposeAndPayAs(buyerAcctName).Test(t).AssertSuccess()
	_ = se.NewTransaction("account_setup_fusd").SignProposeAndPayAs(buyerAcctName).Test(t).AssertSuccess()
	scripts.FundAccountWithFUSD(t, se, buyerAcct.Address(), "1000.0")

	checkDigitalArtNFTSupply(t, se, 0)
	checkDigitalArtCollectionLen(t, se, buyerAcct.Address().String(), 0)

	metadata := &iinft.DigitalArtMetadata{
		MetadataURI: "ipfs://QmMetadata",
		Name:        "Pure Art",
		Artist:      "did:sequel:artist",
		Description: `Digital art in its purest form
The End.`,
		Type:              "Image",
		ContentURI:        "ipfs://QmContent",
		ContentPreviewURI: "ipfs://QmPreview",
		ContentMimetype:   "image/jpeg",
		MaxEdition:        4,
		Asset:             "did:sequel:asset-id",
		Record:            "record-id",
		AssetHead:         "asset-head-id",
	}

	profile := &evergreen.Profile{
		ID: 1,
		Roles: []*evergreen.Role{
			{
				Role:                      evergreen.RoleArtist,
				InitialSaleCommission:     0.9,
				SecondaryMarketCommission: 0.025,
				Address:                   artistAcct.Address(),
			},
			{
				Role:                      evergreen.RolePlatform,
				InitialSaleCommission:     0.05,
				SecondaryMarketCommission: 0.025,
				Address:                   platformAcct.Address(),
			},
			{
				Role:                      "GreenFund",
				InitialSaleCommission:     0.05,
				SecondaryMarketCommission: 0.025,
				Address:                   greenAcct.Address(),
			},
		},
	}

	t.Run("Should be able to mint a token on demand (master not sealed)", func(t *testing.T) {

		_ = client.Transaction(se.GetCustomScript("digitalart_mint_on_demand_fusd", scripts.MintOnDemandParameters{
			Metadata: metadata,
			Profile:  profile,
		})).
			PayloadSigner(buyerAcctName).
			SignProposeAndPayAs(adminAccount).
			StringArgument(metadata.Asset).
			UInt64Argument(1).
			UFix64Argument("100.0").
			UInt64Argument(123).
			Test(t).
			AssertSuccess().
			AssertEventCount(12).
			AssertEmitEventName(
				"A.01cf0e2f2f715450.DigitalArt.Minted",
				"A.01cf0e2f2f715450.DigitalArt.Deposit",
				"A.f8d6e0586b0a20c7.FUSD.TokensWithdrawn",
				"A.f8d6e0586b0a20c7.FUSD.TokensDeposited").
			AssertEmitEvent(gwtf.NewTestEvent("A.01cf0e2f2f715450.DigitalArt.Minted", map[string]interface{}{
				"id":      "0",
				"asset":   "did:sequel:asset-id",
				"edition": "1",
				"modID":   "123",
			})).
			AssertEmitEvent(gwtf.NewTestEvent("A.01cf0e2f2f715450.DigitalArt.Deposit", map[string]interface{}{
				"id": "0",
				"to": "0xe03daebed8ca0615",
			}))

		// Assert that the account's collection is correct
		checkTokenInDigitalArtCollection(t, se, buyerAcct.Address().String(), 0)
		checkDigitalArtCollectionLen(t, se, buyerAcct.Address().String(), 1)
		checkDigitalArtNFTSupply(t, se, 1)

		val, err := se.NewScript("digitalart_get_metadata").
			Argument(cadence.NewAddress(buyerAcct.Address())).
			UInt64Argument(0).
			RunReturns()
		require.NoError(t, err)

		meta, err := iinft.DigitalArtMetadataFromCadence(val)
		require.NoError(t, err)

		assert.Equal(t, uint64(1), meta.Edition)

		assert.Equal(t, 90.0, scripts.GetFUSDBalance(t, se, artistAcct.Address()))
		assert.Equal(t, 900.0, scripts.GetFUSDBalance(t, se, buyerAcct.Address()))
		assert.Equal(t, 5.0, scripts.GetFUSDBalance(t, se, platformAcct.Address()))
		assert.Equal(t, 5.0, scripts.GetFUSDBalance(t, se, greenAcct.Address()))
	})

	t.Run("Should be able to mint a token on demand (master sealed)", func(t *testing.T) {

		_ = client.Transaction(se.GetCustomScript("digitalart_mint_on_demand_fusd", scripts.MintOnDemandParameters{
			Metadata: metadata,
			Profile:  profile,
		})).
			PayloadSigner(buyerAcctName).
			SignProposeAndPayAs(adminAccount).
			StringArgument(metadata.Asset).
			UInt64Argument(1).
			UFix64Argument("100.0").
			UInt64Argument(123).
			Test(t).
			AssertSuccess().
			AssertEventCount(12).
			AssertEmitEventName(
				"A.01cf0e2f2f715450.DigitalArt.Minted",
				"A.01cf0e2f2f715450.DigitalArt.Deposit",
				"A.f8d6e0586b0a20c7.FUSD.TokensWithdrawn",
				"A.f8d6e0586b0a20c7.FUSD.TokensDeposited").
			AssertEmitEvent(gwtf.NewTestEvent("A.01cf0e2f2f715450.DigitalArt.Minted", map[string]interface{}{
				"id":      "1",
				"asset":   "did:sequel:asset-id",
				"edition": "2",
				"modID":   "123",
			})).
			AssertEmitEvent(gwtf.NewTestEvent("A.01cf0e2f2f715450.DigitalArt.Deposit", map[string]interface{}{
				"id": "1",
				"to": "0xe03daebed8ca0615",
			}))

		// Assert that the account's collection is correct
		checkTokenInDigitalArtCollection(t, se, buyerAcct.Address().String(), 1)
		checkDigitalArtCollectionLen(t, se, buyerAcct.Address().String(), 2)
		checkDigitalArtNFTSupply(t, se, 2)

		val, err := se.NewScript("digitalart_get_metadata").
			Argument(cadence.NewAddress(buyerAcct.Address())).
			UInt64Argument(0).
			RunReturns()
		require.NoError(t, err)

		meta, err := iinft.DigitalArtMetadataFromCadence(val)
		require.NoError(t, err)

		assert.Equal(t, uint64(1), meta.Edition)

		assert.Equal(t, 180.0, scripts.GetFUSDBalance(t, se, artistAcct.Address()))
		assert.Equal(t, 800.0, scripts.GetFUSDBalance(t, se, buyerAcct.Address()))
		assert.Equal(t, 10.0, scripts.GetFUSDBalance(t, se, platformAcct.Address()))
		assert.Equal(t, 10.0, scripts.GetFUSDBalance(t, se, greenAcct.Address()))
	})
}

func TestTransferDigitalArt(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true, true)
	require.NoError(t, err)

	scripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	senderAcctName := "emulator-user1"
	senderAcct := client.Account(senderAcctName)

	scripts.FundAccountWithFlow(t, client, senderAcct.Address(), "10.0")

	receiverAcctName := "emulator-user2"
	receiverAcct := client.Account(receiverAcctName)

	scripts.FundAccountWithFlow(t, client, receiverAcct.Address(), "10.0")

	_ = se.NewTransaction("account_setup").
		SignProposeAndPayAs(senderAcctName).
		Test(t).
		AssertSuccess()

	metadata := &iinft.DigitalArtMetadata{
		MetadataURI:       "ipfs://QmMetadata",
		Name:              "Pure Art",
		Artist:            "did:sequel:artist",
		Description:       "Digital art in its purest form",
		Type:              "Image",
		ContentURI:        "ipfs://QmContent",
		ContentPreviewURI: "ipfs://QmPreview",
		ContentMimetype:   "image/jpeg",
		MaxEdition:        4,
		Asset:             "did:sequel:asset-id",
		Record:            "record-id",
		AssetHead:         "asset-head-id",
	}

	profile := &evergreen.Profile{
		ID: 0,
		Roles: []*evergreen.Role{
			{
				Role:                      evergreen.RoleArtist,
				InitialSaleCommission:     0.8,
				SecondaryMarketCommission: 0.2,
				Address:                   senderAcct.Address(),
			},
		},
	}

	_ = scripts.CreateSealDigitalArtTx(se, client, metadata, profile).
		SignProposeAndPayAs(adminAccount).
		Test(t).
		AssertSuccess()

	_ = client.Transaction(se.GetStandardScript("digitalart_mint_edition")).
		SignProposeAndPayAs(adminAccount).
		StringArgument(metadata.Asset).
		UInt64Argument(1).
		Argument(cadence.Address(senderAcct.Address())).
		Test(t).
		AssertSuccess()

	t.Run("Should be able to create a new empty NFT Collection", func(t *testing.T) {

		_ = se.NewTransaction("account_setup").
			SignProposeAndPayAs(receiverAcctName).
			Test(t).
			AssertSuccess()

		checkDigitalArtCollectionLen(t, se, receiverAcct.Address().String(), 0)
	})

	t.Run("Shouldn't be able to withdraw an NFT that doesn't exist in a collection", func(t *testing.T) {

		_ = se.NewTransaction("digitalart_transfer").
			SignProposeAndPayAs(senderAcctName).
			UInt64Argument(3).
			Argument(cadence.Address(receiverAcct.Address())).
			Test(t).
			AssertFailure("missing NFT")

		checkDigitalArtCollectionLen(t, se, receiverAcct.Address().String(), 0)
		checkDigitalArtCollectionLen(t, se, senderAcct.Address().String(), 1)
	})

	t.Run("Should be able to withdraw an NFT and deposit to another accounts collection", func(t *testing.T) {
		_ = se.NewTransaction("digitalart_transfer").
			SignProposeAndPayAs(senderAcctName).
			UInt64Argument(0).
			Argument(cadence.Address(receiverAcct.Address())).
			Test(t).
			AssertSuccess()

		// Assert that the account's collection is correct
		checkTokenInDigitalArtCollection(t, se, receiverAcct.Address().String(), 0)
		checkDigitalArtCollectionLen(t, se, receiverAcct.Address().String(), 1)
		checkDigitalArtCollectionLen(t, se, senderAcct.Address().String(), 0)
	})

	t.Run("Should be able to withdraw an NFT and destroy it, not reducing the supply", func(t *testing.T) {

		_ = se.NewTransaction("digitalart_destroy").
			SignProposeAndPayAs(receiverAcctName).
			UInt64Argument(0).
			Test(t).
			AssertSuccess()

		checkDigitalArtCollectionLen(t, se, receiverAcct.Address().String(), 0)
		checkDigitalArtCollectionLen(t, se, senderAcct.Address().String(), 0)
		checkDigitalArtNFTSupply(t, se, 1)
	})
}

func checkDigitalArtNFTSupply(t *testing.T, se *scripts.Engine, expectedSupply int) {
	_, err := se.NewInlineScript(
		inspectNFTSupplyScript(se.WellKnownAddresses(), "DigitalArt", expectedSupply),
	).RunReturns()
	require.NoError(t, err)
}

func checkTokenInDigitalArtCollection(t *testing.T, se *scripts.Engine, userAddr string, nftID uint64) {
	_, err := se.NewInlineScript(
		inspectCollectionScript(se.WellKnownAddresses(), userAddr, "DigitalArt", "DigitalArt.CollectionPublicPath", nftID),
	).RunReturns()
	require.NoError(t, err)
}

func checkDigitalArtCollectionLen(t *testing.T, se *scripts.Engine, userAddr string, length int) {
	_, err := se.NewInlineScript(
		inspectCollectionLenScript(se.WellKnownAddresses(), userAddr, "DigitalArt", "DigitalArt.CollectionPublicPath", length),
	).RunReturns()
	require.NoError(t, err)
}
