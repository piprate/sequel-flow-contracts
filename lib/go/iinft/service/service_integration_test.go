// +build integration_testing

package service_test

import (
	"testing"

	"github.com/piprate/json-gold/ld"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	. "github.com/piprate/sequel-flow-contracts/lib/go/iinft/service"
	"github.com/stretchr/testify/require"
)

func TestSequelFlowService_MintDigitalArtEdition_Integration(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", false)
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

	editions, err := srv.MintDigitalArtEdition(metadata.Asset, 1, recipientAcct.Address())
	require.NoError(t, err)

	ld.PrintDocument("EDITIONS", editions)
}
