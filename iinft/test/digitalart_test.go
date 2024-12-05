package test

import (
	"context"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/piprate/sequel-flow-contracts/iinft"
	"github.com/piprate/sequel-flow-contracts/iinft/evergreen"
	"github.com/piprate/sequel-flow-contracts/iinft/testscripts"
	"github.com/piprate/splash"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	adminAccountName    = "emulator-sequel-admin"
	platformAccountName = "emulator-sequel-platform"
	user1AccountName    = "emulator-user1"
	user2AccountName    = "emulator-user2"
	user3AccountName    = "emulator-user3"
	initialFlowBalance  = 0.001
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp})
}

func TestDigitalArt_Master(t *testing.T) {
	client, err := splash.NewInMemoryTestConnector("../..", true)
	require.NoError(t, err)

	testscripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := iinft.NewTemplateEngine(client)
	require.NoError(t, err)

	artistAcct := client.Account(user1AccountName)

	metadata := SampleMetadata(2)
	profile, err := evergreen.ProfileToCadence(
		BasicEvergreenProfile(artistAcct.Address), flow.HexToAddress(se.WellKnownAddresses()["Evergreen"]),
	)
	require.NoError(t, err)

	t.Run("Should be able to seal new digital art master", func(t *testing.T) {

		_, err := client.Script(`
import DigitalArt from 0x179b6b1cb6755e31
import Evergreen from 0x179b6b1cb6755e31

access(all) fun main(metadata: DigitalArt.Metadata, evergreenProfile: Evergreen.Profile) {

	// test typical master lifecycle

	var master = DigitalArt.Master(
		metadata: metadata,
		evergreenProfile: evergreenProfile
	)

	assert(master.availableEditions() == 2, message: "wrong availableEditions() value")
	assert(master.newEditionID() == 1, message: "wrong first edition value")
	assert(master.availableEditions() == 1, message: "wrong availableEditions() value")
	assert(master.newEditionID() == 2, message: "wrong first edition value")
	assert(master.availableEditions() == 0, message: "wrong availableEditions() value")

	// this shouldn't happen, but we want to ensure availableEditions() == 0

	assert(master.newEditionID() == 3, message: "wrong first edition value")
	assert(master.availableEditions() == 0, message: "wrong availableEditions() value")

	// close the master

	master.close()
	assert(master.availableEditions() == 0, message: "wrong availableEditions() value")

	// test closing the master before all edition are minted

	master = DigitalArt.Master(
		metadata: metadata,
		evergreenProfile: evergreenProfile
	)

	assert(master.availableEditions() == 2, message: "wrong availableEditions() value")

	master.close()
	assert(master.availableEditions() == 0, message: "wrong availableEditions() value")
}
`).
			Argument(iinft.DigitalArtMetadataToCadence(
				metadata, flow.HexToAddress(se.WellKnownAddresses()["DigitalArt"])),
			).
			Argument(profile).RunReturns(context.Background())
		require.NoError(t, err)
	})
}

func TestDigitalArt_sealMaster(t *testing.T) {
	client, err := splash.NewInMemoryTestConnector("../..", true)
	require.NoError(t, err)

	testscripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := iinft.NewTemplateEngine(client)
	require.NoError(t, err)

	artistAcct := client.Account(user1AccountName)

	profile := BasicEvergreenProfile(artistAcct.Address)

	t.Run("Should be able to seal new digital art master", func(t *testing.T) {

		metadata := SampleMetadata(4)

		_ = testscripts.CreateSealDigitalArtTx(t, se, client, metadata, profile).
			SignProposeAndPayAs(adminAccountName).
			Test(t).
			AssertSuccess()
	})

	t.Run("Shouldn't be able to seal the same digital art master twice", func(t *testing.T) {

		metadata := SampleMetadata(4)
		metadata.Asset = "did:sequel:asset-2"

		// Seal the master

		_ = testscripts.CreateSealDigitalArtTx(t, se, client, metadata, profile).
			SignProposeAndPayAs(adminAccountName).
			Test(t).
			AssertSuccess()

		// try again, should fail
		_ = testscripts.CreateSealDigitalArtTx(t, se, client, metadata, profile).
			SignProposeAndPayAs(adminAccountName).
			Test(t).
			AssertFailure("Master already sealed")
	})

	t.Run("Shouldn't be able to seal the master with an empty asset ID", func(t *testing.T) {

		metadata := SampleMetadata(4)
		metadata.Asset = ""

		_ = testscripts.CreateSealDigitalArtTx(t, se, client, metadata, profile).
			SignProposeAndPayAs(adminAccountName).
			Test(t).
			AssertFailure("Empty asset ID")
	})

	t.Run("Shouldn't be able to seal the master with zero maxEditions", func(t *testing.T) {

		metadata := SampleMetadata(4)
		metadata.MaxEdition = 0

		_ = testscripts.CreateSealDigitalArtTx(t, se, client, metadata, profile).
			SignProposeAndPayAs(adminAccountName).
			Test(t).
			AssertFailure("MaxEdition should be positive")
	})

	t.Run("Shouldn't be able to seal the master with non-zero edition", func(t *testing.T) {

		metadata := SampleMetadata(4)
		metadata.Edition = 2

		_ = testscripts.CreateSealDigitalArtTx(t, se, client, metadata, profile).
			SignProposeAndPayAs(adminAccountName).
			Test(t).
			AssertFailure("Edition should be zero")
	})

	t.Run("Shouldn't be able to re-seal an already closed master", func(t *testing.T) {
		userAcct := client.Account(user1AccountName)

		testscripts.FundAccountWithFlow(t, se, userAcct.Address, "10.0")

		_ = se.NewTransaction("account_setup").
			SignProposeAndPayAs(user1AccountName).
			Test(t).
			AssertSuccess()

		metadata := SampleMetadata(1)
		metadata.Asset = "did:sequel:asset-id-new"

		// seal a master with 1 edition

		_ = testscripts.CreateSealDigitalArtTx(t, se, client, metadata, profile).
			SignProposeAndPayAs(adminAccountName).
			Test(t).
			AssertSuccess()

		// mint the only available edition

		_ = client.Transaction(se.GetStandardScript("digitalart_mint_edition")).
			SignProposeAndPayAs(adminAccountName).
			StringArgument(metadata.Asset).
			UInt64Argument(1).
			Argument(cadence.Address(userAcct.Address)).
			Test(t).
			AssertSuccess()

		// ensure the master is closed

		_, err := client.Script(`
		import DigitalArt from 0x179b6b1cb6755e31

		access(all) fun main(masterId: String) {
			assert(DigitalArt.isClosed(masterId: masterId), message: "master is not closed")
		}
		`).
			StringArgument(metadata.Asset).
			RunReturns(context.Background())
		require.NoError(t, err)

		_ = testscripts.CreateSealDigitalArtTx(t, se, client, metadata, profile).
			SignProposeAndPayAs(adminAccountName).
			Test(t).
			AssertFailure("Master already sealed")
	})
}

func TestDigitalArt_mintEditionNFT(t *testing.T) {
	client, err := splash.NewInMemoryTestConnector("../..", true)
	require.NoError(t, err)

	testscripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := iinft.NewTemplateEngine(client)
	require.NoError(t, err)

	userAcct := client.Account(user1AccountName)

	testscripts.FundAccountWithFlow(t, se, userAcct.Address, "10.0")

	_ = se.NewTransaction("account_setup").
		SignProposeAndPayAs(user1AccountName).
		Test(t).
		AssertSuccess()

	checkDigitalArtNFTSupply(t, se, 0)
	checkDigitalArtCollectionLen(t, se, userAcct.Address.String(), 0)

	metadata := SampleMetadata(2)
	profile := BasicEvergreenProfile(userAcct.Address)

	_ = testscripts.CreateSealDigitalArtTx(t, se, client, metadata, profile).
		SignProposeAndPayAs(adminAccountName).
		Test(t).
		AssertSuccess()

	t.Run("Should be able to mint a token", func(t *testing.T) {

		_ = client.Transaction(se.GetStandardScript("digitalart_mint_edition")).
			SignProposeAndPayAs(adminAccountName).
			StringArgument(metadata.Asset).
			UInt64Argument(1).
			Argument(cadence.Address(userAcct.Address)).
			Test(t).
			AssertSuccess().
			AssertEventCount(8).
			AssertEmitEventName("A.179b6b1cb6755e31.DigitalArt.Minted", "A.179b6b1cb6755e31.DigitalArt.Deposit").
			AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.DigitalArt.Minted", map[string]interface{}{
				"id":      "0",
				"asset":   "did:sequel:asset-id",
				"edition": "1",
				"modID":   "0",
			})).
			AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.DigitalArt.Deposit", map[string]interface{}{
				"id": "0",
				"to": "0xe03daebed8ca0615",
			}))

		// Assert that the account's collection is correct
		checkTokenInDigitalArtCollection(t, se, userAcct.Address.String(), 0)
		checkDigitalArtCollectionLen(t, se, userAcct.Address.String(), 1)
		checkDigitalArtNFTSupply(t, se, 1)

		val, err := se.NewScript("digitalart_get_metadata").
			Argument(cadence.NewAddress(userAcct.Address)).
			UInt64Argument(0).
			RunReturns(context.Background())
		require.NoError(t, err)

		meta, err := iinft.DigitalArtMetadataFromCadence(val)
		require.NoError(t, err)

		assert.Equal(t, uint64(1), meta.Edition)
	})

	t.Run("Editions should have different metadata", func(t *testing.T) {
		_ = client.Transaction(se.GetStandardScript("digitalart_mint_edition")).
			SignProposeAndPayAs(adminAccountName).
			StringArgument(metadata.Asset).
			UInt64Argument(1).
			Argument(cadence.Address(userAcct.Address)).
			Test(t).
			AssertSuccess().
			AssertEventCount(8).
			AssertEmitEventName("A.179b6b1cb6755e31.DigitalArt.Minted", "A.179b6b1cb6755e31.DigitalArt.Deposit").
			AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.DigitalArt.Minted", map[string]interface{}{
				"id":      "1",
				"asset":   "did:sequel:asset-id",
				"edition": "2",
				"modID":   "0",
			})).
			AssertEmitEvent(splash.NewTestEvent("A.179b6b1cb6755e31.DigitalArt.Deposit", map[string]interface{}{
				"id": "1",
				"to": "0xe03daebed8ca0615",
			}))

		// Assert that the account's collection is correct
		checkTokenInDigitalArtCollection(t, se, userAcct.Address.String(), 1)
		checkDigitalArtCollectionLen(t, se, userAcct.Address.String(), 2)
		checkDigitalArtNFTSupply(t, se, 2)

		val, err := se.NewScript("digitalart_get_metadata").
			Argument(cadence.NewAddress(userAcct.Address)).
			UInt64Argument(1).
			RunReturns(context.Background())
		require.NoError(t, err)

		meta, err := iinft.DigitalArtMetadataFromCadence(val)
		require.NoError(t, err)

		assert.Equal(t, uint64(2), meta.Edition)
	})

	t.Run("Should fail if master doesn't exist", func(t *testing.T) {
		_ = client.Transaction(se.GetStandardScript("digitalart_mint_edition")).
			SignProposeAndPayAs(adminAccountName).
			StringArgument("bad_master_id").
			UInt64Argument(1).
			Argument(cadence.Address(userAcct.Address)).
			Test(t).
			AssertFailure("Master not found")
	})

	t.Run("Should fail if no available editions", func(t *testing.T) {
		_ = client.Transaction(`
import DigitalArt from 0x179b6b1cb6755e31

transaction(masterId: String) {
    let admin: &DigitalArt.Admin

    prepare(signer: auth(BorrowValue) &Account) {
        self.admin = signer.storage.borrow<&DigitalArt.Admin>(from: DigitalArt.AdminStoragePath)!
		assert(self.admin.availableEditions(masterId: masterId) == 0, message: "Available editions remain")
    }

    execute {
        let newNFT <- self.admin.mintEditionNFT(masterId: masterId, modID: 0)
		destroy newNFT
    }
}`).
			SignProposeAndPayAs(adminAccountName).
			StringArgument(metadata.Asset).
			Test(t).
			AssertFailure("No more tokens to mint")
	})
}

func TestDigitalArt_isClosed(t *testing.T) {
	client, err := splash.NewInMemoryTestConnector("../..", true)
	require.NoError(t, err)

	testscripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := iinft.NewTemplateEngine(client)
	require.NoError(t, err)

	artistAcct := client.Account(user1AccountName)

	profile := BasicEvergreenProfile(artistAcct.Address)

	t.Run("isClosed() should return false, if master isn't closed", func(t *testing.T) {
		metadata := SampleMetadata(1)

		_ = testscripts.CreateSealDigitalArtTx(t, se, client, metadata, profile).
			SignProposeAndPayAs(adminAccountName).
			Test(t).
			AssertSuccess()

		_, err := client.Script(`
		import DigitalArt from 0x179b6b1cb6755e31

		access(all) fun main(masterId: String) {
			assert(!DigitalArt.isClosed(masterId: masterId), message: "test failed")
		}
		`).
			StringArgument(metadata.Asset).
			RunReturns(context.Background())
		require.NoError(t, err)
	})

	t.Run("isClosed() should return false, if master isn't sealed at all", func(t *testing.T) {
		_, err := client.Script(`
		import DigitalArt from 0x179b6b1cb6755e31

		access(all) fun main(masterId: String) {
			assert(!DigitalArt.isClosed(masterId: masterId), message: "test failed")
		}
		`).
			StringArgument("non-existent-asset").
			RunReturns(context.Background())
		require.NoError(t, err)
	})

	t.Run("isClosed() should return true, if master is closed", func(t *testing.T) {
		userAcct := client.Account(user1AccountName)

		testscripts.FundAccountWithFlow(t, se, userAcct.Address, "10.0")

		_ = se.NewTransaction("account_setup").
			SignProposeAndPayAs(user1AccountName).
			Test(t).
			AssertSuccess()

		metadata := SampleMetadata(1)
		metadata.Asset = "did:sequel:asset-new-id"

		// seal a master with 1 edition

		_ = testscripts.CreateSealDigitalArtTx(t, se, client, metadata, profile).
			SignProposeAndPayAs(adminAccountName).
			Test(t).
			AssertSuccess()

		// mint the only available edition

		_ = client.Transaction(se.GetStandardScript("digitalart_mint_edition")).
			SignProposeAndPayAs(adminAccountName).
			StringArgument(metadata.Asset).
			UInt64Argument(1).
			Argument(cadence.Address(userAcct.Address)).
			Test(t).
			AssertSuccess()

		// ensure the master is closed

		_, err := client.Script(`
		import DigitalArt from 0x179b6b1cb6755e31
		
		access(all) fun main(masterId: String) {
			assert(DigitalArt.isClosed(masterId: masterId), message: "master is not closed")
		}
		`).
			StringArgument(metadata.Asset).
			RunReturns(context.Background())
		require.NoError(t, err)
	})
}

func TestDigitalArt_NFT(t *testing.T) {
	client, err := splash.NewInMemoryTestConnector("../..", true)
	require.NoError(t, err)

	testscripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := iinft.NewTemplateEngine(client)
	require.NoError(t, err)

	userAcct := client.Account(user1AccountName)

	testscripts.FundAccountWithFlow(t, se, userAcct.Address, "10.0")

	_ = se.NewTransaction("account_setup").
		SignProposeAndPayAs(user1AccountName).
		Test(t).
		AssertSuccess()

	metadata := SampleMetadata(4)
	profile := BasicEvergreenProfile(userAcct.Address)
	profile.Roles[0].ReceiverPath = "/public/flowTokenReceiver"
	profile.Roles[0].Description = "artist's royalty"

	_ = testscripts.CreateSealDigitalArtTx(t, se, client, metadata, profile).
		SignProposeAndPayAs(adminAccountName).
		Test(t).
		AssertSuccess()

	_ = client.Transaction(se.GetStandardScript("digitalart_mint_edition")).
		SignProposeAndPayAs(adminAccountName).
		StringArgument(metadata.Asset).
		UInt64Argument(1).
		Argument(cadence.Address(userAcct.Address)).
		Test(t).
		AssertSuccess()

	t.Run("DigitalArt.getMetadata(...) should return NFT metadata", func(t *testing.T) {
		var val cadence.Value
		val, err = se.NewScript("digitalart_get_metadata").
			Argument(cadence.NewAddress(userAcct.Address)).
			UInt64Argument(0).
			RunReturns(context.Background())
		require.NoError(t, err)

		meta, metaErr := iinft.DigitalArtMetadataFromCadence(val)
		require.NoError(t, metaErr)

		assert.Equal(t, metadata.Asset, meta.Asset)
		assert.Equal(t, uint64(1), meta.Edition)
	})

	t.Run("DigitalArt.getMetadata(...) should fail if token doesn't exist in collection", func(t *testing.T) {
		_, err = se.NewScript("digitalart_get_metadata").
			Argument(cadence.NewAddress(userAcct.Address)).
			UInt64Argument(123).
			RunReturns(context.Background())
		require.Error(t, err)
	})

	t.Run("getViews() should return a list of view types", func(t *testing.T) {
		var viewsVal cadence.Value
		viewsVal, err = client.Script(`
import DigitalArt from 0x179b6b1cb6755e31

access(all) fun main(address:Address, tokenID:UInt64) : [Type] {
    let collection = getAccount(address).capabilities.borrow<&{DigitalArt.CollectionPublic}>(DigitalArt.CollectionPublicPath)!
    if let item = collection.borrowDigitalArt(id: tokenID) {
        return item.getViews()
    }

    return []
}
`).
			Argument(cadence.Address(userAcct.Address)).
			UInt64Argument(0).
			RunReturns(context.Background())
		require.NoError(t, err)

		viewsArray, ok := viewsVal.(cadence.Array)
		require.True(t, ok)
		require.Equal(t, 7, len(viewsArray.Values))
		assert.Equal(t, "Type<A.f8d6e0586b0a20c7.MetadataViews.Display>()", viewsArray.Values[0].String())
		assert.Equal(t, "Type<A.f8d6e0586b0a20c7.MetadataViews.Edition>()", viewsArray.Values[1].String())
		assert.Equal(t, "Type<A.f8d6e0586b0a20c7.MetadataViews.Royalties>()", viewsArray.Values[2].String())
		assert.Equal(t, "Type<A.f8d6e0586b0a20c7.MetadataViews.ExternalURL>()", viewsArray.Values[3].String())
		assert.Equal(t, "Type<A.f8d6e0586b0a20c7.MetadataViews.NFTCollectionData>()", viewsArray.Values[4].String())
		assert.Equal(t, "Type<A.f8d6e0586b0a20c7.MetadataViews.NFTCollectionDisplay>()", viewsArray.Values[5].String())
		assert.Equal(t, "Type<A.179b6b1cb6755e31.DigitalArt.Metadata>()", viewsArray.Values[6].String())
	})

	t.Run("resolveView(Type<MetadataViews.Display>()) should return MetadataViews.Display view", func(t *testing.T) {
		var val cadence.Value
		val, err = client.Script(`
import MetadataViews from 0xf8d6e0586b0a20c7
import DigitalArt from 0x179b6b1cb6755e31

access(all) fun main(address:Address, tokenID:UInt64) : MetadataViews.Display? {
    let collection = getAccount(address).capabilities.borrow<&{DigitalArt.CollectionPublic}>(DigitalArt.CollectionPublicPath)!
    if let item = collection.borrowDigitalArt(id: tokenID) {
        if let view = item.resolveView(Type<MetadataViews.Display>()) {
            return view as! MetadataViews.Display
        }
    }

    return nil
}
`).
			Argument(cadence.Address(userAcct.Address)).
			UInt64Argument(0).
			RunReturns(context.Background())
		require.NoError(t, err)

		displayStruct, ok := val.(cadence.Optional).Value.(cadence.Struct)
		require.True(t, ok)
		assert.Equal(t, "MetadataViews.Display", displayStruct.StructType.QualifiedIdentifier)
		assert.Equal(t, cadence.String("Pure Art"), displayStruct.SearchFieldByName("name"))
		assert.Equal(t, cadence.String("Digital art in its purest form"), displayStruct.SearchFieldByName("description"))
		thumbnailStruct, ok := displayStruct.SearchFieldByName("thumbnail").(cadence.Struct)
		require.True(t, ok)
		assert.Equal(t, "MetadataViews.HTTPFile", thumbnailStruct.StructType.QualifiedIdentifier)
		assert.Equal(t, cadence.String("https://sequel.mypinata.cloud/ipfs/QmPreview"), thumbnailStruct.SearchFieldByName("url"))
	})

	t.Run("resolveView(Type<MetadataViews.Royalties>()) should return MetadataViews.Royalties view", func(t *testing.T) {

		_, err = client.Script(`
import MetadataViews from 0xf8d6e0586b0a20c7
import DigitalArt from 0x179b6b1cb6755e31

access(all) fun main(address:Address, tokenID:UInt64) {
    let collection = getAccount(address).capabilities.borrow<&{DigitalArt.CollectionPublic}>(DigitalArt.CollectionPublicPath)!

	var royalties: [MetadataViews.Royalty] = []
	if let item = collection.borrowDigitalArt(id: tokenID) {
        if let view = item.resolveView(Type<MetadataViews.Royalties>()) {
            royalties = (view as! MetadataViews.Royalties).getRoyalties()
        }
    }

	assert(royalties != nil, message: "royalties == nil")
	assert(royalties.length == 1, message: "incorrect number of royalties")

	assert(royalties[0].receiver.check(), message: "bad royalty receiver")
	assert(royalties[0].cut == 0.05, message: "incorrect royalty cut")
	assert(royalties[0].description == "artist's royalty", message: "incorrect royalty description")
}
`).
			Argument(cadence.Address(userAcct.Address)).
			UInt64Argument(0).
			RunReturns(context.Background())
		require.NoError(t, err)
	})

	t.Run("resolveView(Type<MetadataViews.ExternalURL>()) should return MetadataViews.ExternalURL view", func(t *testing.T) {

		_, err = client.Script(`
import MetadataViews from 0xf8d6e0586b0a20c7
import DigitalArt from 0x179b6b1cb6755e31

access(all) fun main(address:Address, tokenID:UInt64) {
    let collection = getAccount(address).capabilities.borrow<&{DigitalArt.CollectionPublic}>(DigitalArt.CollectionPublicPath)!

	var externalURL: MetadataViews.ExternalURL? = nil
	if let item = collection.borrowDigitalArt(id: tokenID) {
        if let view = item.resolveView(Type<MetadataViews.ExternalURL>()) {
            externalURL = (view as! MetadataViews.ExternalURL)
        }
    }

	assert(externalURL != nil, message: "externalURL == nil")
	assert(externalURL!.url == "https://app.sequel.space/tokens/digital-art/0", message: "incorrect external URL")
}
`).
			Argument(cadence.Address(userAcct.Address)).
			UInt64Argument(0).
			RunReturns(context.Background())
		require.NoError(t, err)
	})

	t.Run("resolveView(Type<DigitalArt.Metadata>()) should return DigitalArt.Metadata view", func(t *testing.T) {
		var val cadence.Value
		val, err = client.Script(`
import MetadataViews from 0xf8d6e0586b0a20c7
import DigitalArt from 0x179b6b1cb6755e31

access(all) fun main(address:Address, tokenID:UInt64) : DigitalArt.Metadata? {
    let collection = getAccount(address).capabilities.borrow<&{DigitalArt.CollectionPublic}>(DigitalArt.CollectionPublicPath)!
    if let item = collection.borrowDigitalArt(id: tokenID) {
        if let view = item.resolveView(Type<DigitalArt.Metadata>()) {
            return view as! DigitalArt.Metadata
        }
    }

    return nil
}
`).
			Argument(cadence.Address(userAcct.Address)).
			UInt64Argument(0).
			RunReturns(context.Background())
		require.NoError(t, err)

		meta, metaErr := iinft.DigitalArtMetadataFromCadence(val)
		require.NoError(t, metaErr)

		assert.Equal(t, "Pure Art", meta.Name)
		assert.Equal(t, "Digital art in its purest form", meta.Description)
		assert.Equal(t, uint64(1), meta.Edition)
	})

	t.Run("getAssetID() should return DigitalArt's master ID", func(t *testing.T) {
		var val cadence.Value
		val, err = client.Script(`
import DigitalArt from 0x179b6b1cb6755e31

access(all) fun main(address:Address, tokenID:UInt64) : String {
    let collection = getAccount(address).capabilities.borrow<&{DigitalArt.CollectionPublic}>(DigitalArt.CollectionPublicPath)!
    if let item = collection.borrowDigitalArt(id: tokenID) {
        return item.getAssetID()
    }

    return ""
}
`).
			Argument(cadence.Address(userAcct.Address)).
			UInt64Argument(0).
			RunReturns(context.Background())
		require.NoError(t, err)

		assert.Equal(t, cadence.String(metadata.Asset), val)
	})

	t.Run("getEvergreenProfile() should return DigitalArt's Evergreen profile", func(t *testing.T) {
		var val cadence.Value
		val, err = client.Script(`
import DigitalArt from 0x179b6b1cb6755e31
import Evergreen from 0x179b6b1cb6755e31

access(all) fun main(address:Address, tokenID:UInt64) : Evergreen.Profile? {
    let collection = getAccount(address).capabilities.borrow<&{DigitalArt.CollectionPublic}>(DigitalArt.CollectionPublicPath)!
    if let item = collection.borrowDigitalArt(id: tokenID) {
        return item.getEvergreenProfile()
    }

    return nil
}
`).
			Argument(cadence.Address(userAcct.Address)).
			UInt64Argument(0).
			RunReturns(context.Background())
		require.NoError(t, err)

		actual, err := evergreen.ProfileFromCadence(val)
		require.NoError(t, err)
		assert.Equal(t, profile.ID, actual.ID)
	})
}

func TestDigitalArt_Collection(t *testing.T) {
	client, err := splash.NewInMemoryTestConnector("../..", true)
	require.NoError(t, err)

	testscripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := iinft.NewTemplateEngine(client)
	require.NoError(t, err)

	userAcct := client.Account(user1AccountName)

	testscripts.FundAccountWithFlow(t, se, userAcct.Address, "10.0")

	_ = se.NewTransaction("account_setup").
		SignProposeAndPayAs(user1AccountName).
		Test(t).
		AssertSuccess()

	metadata := SampleMetadata(4)
	profile := BasicEvergreenProfile(userAcct.Address)

	_ = testscripts.CreateSealDigitalArtTx(t, se, client, metadata, profile).
		SignProposeAndPayAs(adminAccountName).
		Test(t).
		AssertSuccess()

	// mint 2 NFTs

	_ = client.Transaction(se.GetStandardScript("digitalart_mint_edition")).
		SignProposeAndPayAs(adminAccountName).
		StringArgument(metadata.Asset).
		UInt64Argument(2).
		Argument(cadence.Address(userAcct.Address)).
		Test(t).
		AssertSuccess()

	t.Run("getIDs() should return a list of token IDs", func(t *testing.T) {
		var viewsVal cadence.Value
		viewsVal, err = client.Script(`
import DigitalArt from 0x179b6b1cb6755e31

access(all) fun main(address:Address, tokenID:UInt64) : [UInt64] {
    let collection = getAccount(address).capabilities.borrow<&{DigitalArt.CollectionPublic}>(DigitalArt.CollectionPublicPath)!
    return collection.getIDs()
}
`).
			Argument(cadence.Address(userAcct.Address)).
			UInt64Argument(0).
			RunReturns(context.Background())
		require.NoError(t, err)

		idArray, ok := viewsVal.(cadence.Array)
		require.True(t, ok)
		require.Equal(t, 2, len(idArray.Values))
		ids := []uint64{
			uint64(idArray.Values[0].(cadence.UInt64)),
			uint64(idArray.Values[1].(cadence.UInt64)),
		}
		slices.Sort(ids)
		assert.Equal(t, uint64(0), ids[0])
		assert.Equal(t, uint64(1), ids[1])
	})

	t.Run("borrowNFT(...) should return NonFungibleToken.NFT", func(t *testing.T) {
		var val cadence.Value
		val, err = client.Script(`
import NonFungibleToken from 0xf8d6e0586b0a20c7
import DigitalArt from 0x179b6b1cb6755e31

access(all) fun main(address:Address, tokenID:UInt64) : UInt64 {
	let collection = getAccount(address).capabilities.borrow<&{NonFungibleToken.CollectionPublic}>(DigitalArt.CollectionPublicPath)!
	let tokenRef = collection.borrowNFT(tokenID) as! &DigitalArt.NFT?

	return tokenRef!.id
}
`).
			Argument(cadence.Address(userAcct.Address)).
			UInt64Argument(1).
			RunReturns(context.Background())
		require.NoError(t, err)

		assert.Equal(t, cadence.UInt64(1), val)
	})

	t.Run("borrowNFT(...) should fail if NFT isn't found", func(t *testing.T) {

		_, err = client.Script(`
import NonFungibleToken from 0xf8d6e0586b0a20c7
import DigitalArt from 0x179b6b1cb6755e31

access(all) fun main(address:Address, tokenID:UInt64) : UInt64 {
    let collection = getAccount(address).capabilities.borrow<&{NonFungibleToken.CollectionPublic}>(DigitalArt.CollectionPublicPath)!
	let tokenRef = collection.borrowNFT(tokenID)

	return tokenRef.id
}
`).
			Argument(cadence.Address(userAcct.Address)).
			UInt64Argument(2).
			RunReturns(context.Background())
		require.Error(t, err)
	})

	t.Run("borrowDigitalArt(...) should return DigitalArt.NFT", func(t *testing.T) {
		var val cadence.Value
		val, err = client.Script(`
import DigitalArt from 0x179b6b1cb6755e31

access(all) fun main(address:Address, tokenID:UInt64) : String {
    let collection = getAccount(address).capabilities.borrow<&{DigitalArt.CollectionPublic}>(DigitalArt.CollectionPublicPath)!
    let daToken = collection.borrowDigitalArt(id: tokenID)
	return daToken!.metadata.asset
}
`).
			Argument(cadence.Address(userAcct.Address)).
			UInt64Argument(1).
			RunReturns(context.Background())
		require.NoError(t, err)

		assert.Equal(t, cadence.String(metadata.Asset), val)
	})

	t.Run("borrowDigitalArt(...) should return nil if token isn't found", func(t *testing.T) {
		var val cadence.Value
		val, err = client.Script(`
import DigitalArt from 0x179b6b1cb6755e31

access(all) fun main(address:Address, tokenID:UInt64) : String {
    let collection = getAccount(address).capabilities.borrow<&{DigitalArt.CollectionPublic}>(DigitalArt.CollectionPublicPath)!
    if let item = collection.borrowDigitalArt(id: tokenID) {
		return item.metadata.asset
	}
	return "not found"
}
`).
			Argument(cadence.Address(userAcct.Address)).
			UInt64Argument(123).
			RunReturns(context.Background())
		require.NoError(t, err)

		assert.Equal(t, cadence.String("not found"), val)
	})

	t.Run("borrowViewResolver(...) should return MetadataViews.Display view", func(t *testing.T) {

		_, err = client.Script(`
import MetadataViews from 0xf8d6e0586b0a20c7
import ViewResolver from 0xf8d6e0586b0a20c7
import DigitalArt from 0x179b6b1cb6755e31

access(all) fun main(address:Address, tokenID:UInt64) {
	let collection = getAccount(address).capabilities.borrow<&{ViewResolver.ResolverCollection}>(DigitalArt.CollectionPublicPath)!
	
	let resolver = collection.borrowViewResolver(id: tokenID)!

	if let view = resolver.resolveView(Type<MetadataViews.Display>()) {
		let display = view as! MetadataViews.Display
		assert(display.name == "Pure Art", message: "bad value of meta.name")
	} else {
		panic("MetadataViews.Display view not found")
	}

	if let view = resolver.resolveView(Type<DigitalArt.Metadata>()) {
		let meta = view as! DigitalArt.Metadata
		assert(meta.name == "Pure Art", message: "bad value of meta.name")
	} else {
		panic("DigitalArt.Metadata view not found")
	}

}
`).
			Argument(cadence.Address(userAcct.Address)).
			UInt64Argument(0).
			RunReturns(context.Background())
		require.NoError(t, err)
	})

	t.Run("borrowEvergreenToken(...) should return &AnyResource{Evergreen.Token}", func(t *testing.T) {

		_, err = client.Script(`
import MetadataViews from 0xf8d6e0586b0a20c7
import Evergreen from 0x179b6b1cb6755e31
import DigitalArt from 0x179b6b1cb6755e31

access(all) fun main(address:Address, tokenID:UInt64) {
	let collection = getAccount(address).capabilities.borrow<&{Evergreen.CollectionPublic}>(DigitalArt.CollectionPublicPath)!
	
	let token = collection.borrowEvergreenToken(id: tokenID)

	let profile = token!.getEvergreenProfile()

	assert(profile.id == "did:sequel:evergreen1", message: "bad value of evergreen profile ID")
}
`).
			Argument(cadence.Address(userAcct.Address)).
			UInt64Argument(0).
			RunReturns(context.Background())
		require.NoError(t, err)
	})

	t.Run("borrowEvergreenToken(...) should return nil if token not found", func(t *testing.T) {

		_, err = client.Script(`
import MetadataViews from 0xf8d6e0586b0a20c7
import Evergreen from 0x179b6b1cb6755e31
import DigitalArt from 0x179b6b1cb6755e31

access(all) fun main(address:Address, tokenID:UInt64) {
	let collection = getAccount(address).capabilities.borrow<&{Evergreen.CollectionPublic}>(DigitalArt.CollectionPublicPath)!

	let token = collection.borrowEvergreenToken(id: tokenID)

	assert(token == nil, message: "token not nil")
}
`).
			Argument(cadence.Address(userAcct.Address)).
			UInt64Argument(123).
			RunReturns(context.Background())
		require.NoError(t, err)
	})
}

func TestDigitalArt_createEmptyCollection(t *testing.T) {
	client, err := splash.NewInMemoryTestConnector("../..", true)
	require.NoError(t, err)

	testscripts.ConfigureInMemoryEmulator(t, client, "1000.0")

	se, err := iinft.NewTemplateEngine(client)
	require.NoError(t, err)

	receiverAcctName := user2AccountName
	receiverAcct := client.Account(receiverAcctName)

	testscripts.FundAccountWithFlow(t, se, receiverAcct.Address, "10.0")

	t.Run("Should be able to create a new empty NFT Collection", func(t *testing.T) {

		_ = se.NewTransaction("account_setup").
			SignProposeAndPayAs(receiverAcctName).
			Test(t).
			AssertSuccess()

		checkDigitalArtCollectionLen(t, se, receiverAcct.Address.String(), 0)
	})
}

func checkDigitalArtNFTSupply(t *testing.T, se *splash.TemplateEngine, expectedSupply int) {
	t.Helper()

	_, err := se.NewInlineScript(
		inspectNFTSupplyScript(se.WellKnownAddresses(), "DigitalArt", expectedSupply),
	).RunReturns(context.Background())
	require.NoError(t, err)
}

func checkTokenInDigitalArtCollection(t *testing.T, se *splash.TemplateEngine, userAddr string, nftID uint64) {
	t.Helper()

	_, err := se.NewInlineScript(
		inspectCollectionScript(se.WellKnownAddresses(), userAddr, "DigitalArt", "DigitalArt.CollectionPublicPath", nftID),
	).RunReturns(context.Background())
	require.NoError(t, err)
}

func checkDigitalArtCollectionLen(t *testing.T, se *splash.TemplateEngine, userAddr string, length int) {
	t.Helper()

	_, err := se.NewInlineScript(
		inspectCollectionLenScript(se.WellKnownAddresses(), userAddr, "DigitalArt", "DigitalArt.CollectionPublicPath", length),
	).RunReturns(context.Background())
	require.NoError(t, err)
}
