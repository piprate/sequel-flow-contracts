package test

import (
	"os"
	"testing"
	"time"

	"github.com/onflow/cadence"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/gwtf"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/scripts"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sequelAccount = "account"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp})
}

func TestSealDigitalArtMaster(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true)
	require.NoError(t, err)

	client.InitializeContracts()

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	userAcct, err := client.State.Accounts().ByName("emulator-user1")
	require.NoError(t, err)

	sampleMetadata := &iinft.Metadata{
		MetadataLink:       "QmMetadata",
		Name:               "Pure Art",
		Artist:             "Arty",
		Description:        "Digital art in its purest form",
		Type:               "Image",
		ContentLink:        "QmContent",
		ContentPreviewLink: "QmPreview",
		Mimetype:           "image/jpeg",
		MaxEdition:         4,
		Asset:              "did:sequel:asset-id",
		Record:             "record-id",
		AssetHead:          "asset-head-id",
		ParticipationProfile: &iinft.ParticipationProfile{
			ID: 0,
			Roles: map[string]*iinft.ParticipationRole{
				iinft.ParticipationRoleArtist: {
					Role:                      iinft.ParticipationRoleArtist,
					InitialSaleCommission:     80.0,
					SecondaryMarketCommission: 20.0,
					Address:                   userAcct.Address(),
				},
			},
		},
	}

	t.Run("Should be able to seal new digital art master", func(t *testing.T) {

		_ = scripts.CreateSealDigitalArtTx(se.GetStandardScript("master_seal"), client, sampleMetadata).
			SignProposeAndPayAs(sequelAccount).
			Test(t).
			AssertSuccess()
	})

	t.Run("Shouldn't be able to seal the same digital art master twice", func(t *testing.T) {

		sampleMetadata2 := *sampleMetadata
		sampleMetadata2.Asset = "did:sequel:asset-2"

		// Seal the master
		script := se.GetStandardScript("master_seal")

		_ = scripts.CreateSealDigitalArtTx(script, client, &sampleMetadata2).
			SignProposeAndPayAs(sequelAccount).
			Test(t).
			AssertSuccess()

		// try again
		_ = scripts.CreateSealDigitalArtTx(script, client, &sampleMetadata2).
			SignProposeAndPayAs(sequelAccount).
			Test(t).
			AssertFailure("master already sealed")
	})
}

func TestMintDigitalArtEditions(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true)
	require.NoError(t, err)

	client.InitializeContracts().CreateAccounts("emulator-account")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	userAcct, err := client.State.Accounts().ByName("emulator-user1")
	require.NoError(t, err)

	_ = se.NewTransaction("account_setup").
		SignProposeAndPayAs("user1").
		Test(t).
		AssertSuccess()

	checkDigitalArtNFTSupply(t, se, 0)
	checkDigitalArtCollectionLen(t, se, userAcct.Address().String(), 0)

	metadata := &iinft.Metadata{
		MetadataLink:       "QmMetadata",
		Name:               "Pure Art",
		Artist:             "Arty",
		Description:        "Digital art in its purest form",
		Type:               "Image",
		ContentLink:        "QmContent",
		ContentPreviewLink: "QmPreview",
		Mimetype:           "image/jpeg",
		MaxEdition:         4,
		Asset:              "did:sequel:asset-id",
		Record:             "record-id",
		AssetHead:          "asset-head-id",
		ParticipationProfile: &iinft.ParticipationProfile{
			ID: 0,
			Roles: map[string]*iinft.ParticipationRole{
				iinft.ParticipationRoleArtist: {
					Role:                      iinft.ParticipationRoleArtist,
					InitialSaleCommission:     80.0,
					SecondaryMarketCommission: 20.0,
					Address:                   userAcct.Address(),
				},
			},
		},
	}

	_ = scripts.CreateSealDigitalArtTx(se.GetStandardScript("master_seal"), client, metadata).
		SignProposeAndPayAs(sequelAccount).
		Test(t).
		AssertSuccess()

	t.Run("Should be able to mint a token", func(t *testing.T) {

		_ = client.Transaction(se.GetStandardScript("digitalart_mint_edition")).
			SignProposeAndPayAs(sequelAccount).
			StringArgument(metadata.Asset).
			UInt64Argument(1).
			Argument(cadence.Address(userAcct.Address())).
			Test(t).
			AssertSuccess().
			AssertEventCount(2).
			AssertEmitEventName("A.f8d6e0586b0a20c7.DigitalArt.Minted", "A.f8d6e0586b0a20c7.DigitalArt.Deposit").
			AssertEmitEvent(gwtf.NewTestEvent("A.f8d6e0586b0a20c7.DigitalArt.Minted", map[string]interface{}{
				"id":      "0",
				"asset":   "did:sequel:asset-id",
				"edition": "1",
			})).
			AssertEmitEvent(gwtf.NewTestEvent("A.f8d6e0586b0a20c7.DigitalArt.Deposit", map[string]interface{}{
				"id": "0",
				"to": "0x1cf0e2f2f715450",
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

		meta, err := iinft.NewMetadataFromCadence(val)
		require.NoError(t, err)

		assert.Equal(t, uint64(1), meta.Edition)
	})

	t.Run("Editions should have different metadata", func(t *testing.T) {
		_ = client.Transaction(se.GetStandardScript("digitalart_mint_edition")).
			SignProposeAndPayAs(sequelAccount).
			StringArgument(metadata.Asset).
			UInt64Argument(1).
			Argument(cadence.Address(userAcct.Address())).
			Test(t).
			AssertSuccess().
			AssertEventCount(2).
			AssertEmitEventName("A.f8d6e0586b0a20c7.DigitalArt.Minted", "A.f8d6e0586b0a20c7.DigitalArt.Deposit").
			AssertEmitEvent(gwtf.NewTestEvent("A.f8d6e0586b0a20c7.DigitalArt.Minted", map[string]interface{}{
				"id":      "1",
				"asset":   "did:sequel:asset-id",
				"edition": "2",
			})).
			AssertEmitEvent(gwtf.NewTestEvent("A.f8d6e0586b0a20c7.DigitalArt.Deposit", map[string]interface{}{
				"id": "1",
				"to": "0x1cf0e2f2f715450",
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

		meta, err := iinft.NewMetadataFromCadence(val)
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

func TestMintDigitalArtSingles(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true)
	require.NoError(t, err)

	client.InitializeContracts().CreateAccounts("emulator-account")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	userAcct, err := client.State.Accounts().ByName("emulator-user1")
	require.NoError(t, err)

	_ = se.NewTransaction("account_setup").
		SignProposeAndPayAs("user1").
		Test(t).
		AssertSuccess()

	checkDigitalArtNFTSupply(t, se, 0)
	checkDigitalArtCollectionLen(t, se, userAcct.Address().String(), 0)

	metadata := &iinft.Metadata{
		MetadataLink:       "QmMetadata",
		Name:               "Pure Art",
		Artist:             "Arty",
		Description:        "Digital art in its purest form",
		Type:               "Image",
		ContentLink:        "QmContent",
		ContentPreviewLink: "QmPreview",
		Mimetype:           "image/jpeg",
		Asset:              "did:sequel:asset-id",
		Record:             "record-id",
		AssetHead:          "asset-head-id",
		ParticipationProfile: &iinft.ParticipationProfile{
			ID: 0,
			Roles: map[string]*iinft.ParticipationRole{
				iinft.ParticipationRoleArtist: {
					Role:                      iinft.ParticipationRoleArtist,
					InitialSaleCommission:     80.0,
					SecondaryMarketCommission: 20.0,
					Address:                   userAcct.Address(),
				},
			},
		},
	}

	t.Run("Should be able to mint a token", func(t *testing.T) {

		_ = scripts.CreateMintSingleDigitalArtTx(se.GetStandardScript("digitalart_mint_single"), client, metadata, userAcct.Address()).
			SignProposeAndPayAs(sequelAccount).
			Test(t).
			AssertSuccess().
			AssertEventCount(2).
			AssertEmitEventName("A.f8d6e0586b0a20c7.DigitalArt.Minted", "A.f8d6e0586b0a20c7.DigitalArt.Deposit").
			AssertEmitEvent(gwtf.NewTestEvent("A.f8d6e0586b0a20c7.DigitalArt.Minted", map[string]interface{}{
				"id":      "0",
				"asset":   "did:sequel:asset-id",
				"edition": "1",
			})).
			AssertEmitEvent(gwtf.NewTestEvent("A.f8d6e0586b0a20c7.DigitalArt.Deposit", map[string]interface{}{
				"id": "0",
				"to": "0x1cf0e2f2f715450",
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

		meta, err := iinft.NewMetadataFromCadence(val)
		require.NoError(t, err)

		assert.Equal(t, uint64(1), meta.Edition)
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

func TestTransferDigitalArt(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true)
	require.NoError(t, err)

	client.InitializeContracts().CreateAccounts("emulator-account")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	senderAccount := "user1"
	senderAcct, err := client.State.Accounts().ByName("emulator-" + senderAccount)
	require.NoError(t, err)

	receiverAccount := "user2"
	receiverAcct, err := client.State.Accounts().ByName("emulator-" + receiverAccount)
	require.NoError(t, err)

	_ = se.NewTransaction("account_setup").
		SignProposeAndPayAs(senderAccount).
		Test(t).
		AssertSuccess()

	metadata := &iinft.Metadata{
		MetadataLink:       "QmMetadata",
		Name:               "Pure Art",
		Artist:             "Arty",
		Description:        "Digital art in its purest form",
		Type:               "Image",
		ContentLink:        "QmContent",
		ContentPreviewLink: "QmPreview",
		Mimetype:           "image/jpeg",
		MaxEdition:         4,
		Asset:              "did:sequel:asset-id",
		Record:             "record-id",
		AssetHead:          "asset-head-id",
		ParticipationProfile: &iinft.ParticipationProfile{
			ID: 0,
			Roles: map[string]*iinft.ParticipationRole{
				iinft.ParticipationRoleArtist: {
					Role:                      iinft.ParticipationRoleArtist,
					InitialSaleCommission:     80.0,
					SecondaryMarketCommission: 20.0,
					Address:                   senderAcct.Address(),
				},
			},
		},
	}

	_ = scripts.CreateSealDigitalArtTx(se.GetStandardScript("master_seal"), client, metadata).
		SignProposeAndPayAs(sequelAccount).
		Test(t).
		AssertSuccess()

	_ = client.Transaction(se.GetStandardScript("digitalart_mint_edition")).
		SignProposeAndPayAs(sequelAccount).
		StringArgument(metadata.Asset).
		UInt64Argument(1).
		Argument(cadence.Address(senderAcct.Address())).
		Test(t).
		AssertSuccess()

	t.Run("Should be able to create a new empty NFT Collection", func(t *testing.T) {

		_ = se.NewTransaction("account_setup").
			SignProposeAndPayAs(receiverAccount).
			Test(t).
			AssertSuccess()

		checkDigitalArtCollectionLen(t, se, receiverAcct.Address().String(), 0)
	})

	t.Run("Shouldn't be able to withdraw an NFT that doesn't exist in a collection", func(t *testing.T) {

		_ = se.NewTransaction("digitalart_transfer").
			SignProposeAndPayAs(senderAccount).
			UInt64Argument(3).
			Argument(cadence.Address(receiverAcct.Address())).
			Test(t).
			AssertFailure("missing NFT")

		checkDigitalArtCollectionLen(t, se, receiverAcct.Address().String(), 0)
		checkDigitalArtCollectionLen(t, se, senderAcct.Address().String(), 1)
	})

	t.Run("Should be able to withdraw an NFT and deposit to another accounts collection", func(t *testing.T) {
		_ = se.NewTransaction("digitalart_transfer").
			SignProposeAndPayAs(senderAccount).
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
			SignProposeAndPayAs(receiverAccount).
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
