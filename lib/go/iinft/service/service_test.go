package service_test

import (
	"os"
	"testing"
	"time"

	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	. "github.com/piprate/sequel-flow-contracts/lib/go/iinft/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp})
}

func TestSequelFlowService_SealDigitalArt(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true)
	require.NoError(t, err)

	client.InitializeContracts()

	sequelAcct, err := client.State.Accounts().ByName("emulator-account")
	require.NoError(t, err)

	srv, err := NewSequelFlowService(client, sequelAcct)
	require.NoError(t, err)

	userAcct, err := client.State.Accounts().ByName("emulator-user1")
	require.NoError(t, err)

	metadata := &iinft.Metadata{
		MetadataLink:       "QmMetadata",
		Name:               "Pure Art",
		Artist:             "Arty",
		ArtistAddress:      userAcct.Address(),
		Description:        "Digital art in its purest form",
		Type:               "Image",
		ContentLink:        "QmContent",
		ContentPreviewLink: "QmPreview",
		Mimetype:           "image/jpeg",
		MaxEdition:         4,
		Asset:              "did:sequel:asset-id",
		Record:             "record-id",
		AssetHead:          "asset-head-id",
	}

	t.Run("Should be able to seal new digital art master", func(t *testing.T) {
		err = srv.SealDigitalArt(metadata)
		require.NoError(t, err)
	})

	t.Run("Shouldn't be able to seal the same digital art master twice", func(t *testing.T) {
		metadata2 := *metadata
		metadata2.Asset = "did:sequel:asset-2"

		err = srv.SealDigitalArt(&metadata2)
		require.NoError(t, err)

		// try again
		err = srv.SealDigitalArt(&metadata2)
		require.Error(t, err)
	})
}

func TestSequelFlowService_MintDigitalArtEdition(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true)
	require.NoError(t, err)

	client.InitializeContracts().CreateAccounts("emulator-account")

	sequelAcct, err := client.State.Accounts().ByName("emulator-account")
	require.NoError(t, err)

	srv, err := NewSequelFlowService(client, sequelAcct)
	require.NoError(t, err)

	userAcct, err := client.State.Accounts().ByName("emulator-user1")
	require.NoError(t, err)

	metadata := &iinft.Metadata{
		MetadataLink:       "QmMetadata",
		Name:               "Pure Art",
		Artist:             "Arty",
		ArtistAddress:      userAcct.Address(),
		Description:        "Digital art in its purest form",
		Type:               "Image",
		ContentLink:        "QmContent",
		ContentPreviewLink: "QmPreview",
		Mimetype:           "image/jpeg",
		MaxEdition:         4,
		Asset:              "did:sequel:asset-id",
		Record:             "record-id",
		AssetHead:          "asset-head-id",
	}

	err = srv.SealDigitalArt(metadata)
	require.NoError(t, err)

	recipientAcct, err := client.State.Accounts().ByName("emulator-user2")
	require.NoError(t, err)

	err = srv.CreateDigitalArtCollection(recipientAcct)
	require.NoError(t, err)

	t.Run("Should be able to mint a token", func(t *testing.T) {
		editions, err := srv.MintDigitalArtEdition(metadata.Asset, 1, recipientAcct.Address())
		require.NoError(t, err)

		if assert.Equal(t, 1, len(editions)) {
			assert.Equal(t, 0, editions[0].ID)
			assert.Equal(t, metadata.Asset, editions[0].Asset)
			assert.Equal(t, 1, editions[0].EditionNumber)
		}
	})

	t.Run("Should be able to mint multiple tokens", func(t *testing.T) {
		editions, err := srv.MintDigitalArtEdition(metadata.Asset, 2, recipientAcct.Address())
		require.NoError(t, err)

		if assert.Equal(t, 2, len(editions)) {
			assert.Equal(t, 1, editions[0].ID)
			assert.Equal(t, metadata.Asset, editions[0].Asset)
			assert.Equal(t, 2, editions[0].EditionNumber)
			assert.Equal(t, 2, editions[1].ID)
			assert.Equal(t, metadata.Asset, editions[1].Asset)
			assert.Equal(t, 3, editions[1].EditionNumber)
		}
	})

	t.Run("Should be not able to mint more tokens than available", func(t *testing.T) {
		_, err = srv.MintDigitalArtEdition(metadata.Asset, 2, recipientAcct.Address())
		require.Error(t, err)
	})
}
