package iinft

import (
	"errors"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/common"
	"github.com/onflow/flow-go-sdk"
	"github.com/piprate/sequel-flow-contracts/iinft/evergreen"
)

type (
	DigitalArtMetadata struct {
		// Asset is the DID of the master's asset.
		// This ID is the same for all editions of a particular Digital Art NFT.
		Asset string
		// Name is the name of the digital art.
		//
		// This field will be displayed in lists and therefore should
		// be short and concise.
		//
		Name string
		// Artist is the DID of the artist who created the given digital art.
		Artist string
		// Description is a written description of the digital art.
		//
		// This field will be displayed in a detailed view of the object,
		// so can be more verbose (e.g. a paragraph instead of a single line).
		//
		Description string
		// Type is the digital art's media type: Image, Audio, Video, etc.
		Type string
		// ContentURI is the URI of the original digital art content.
		ContentURI string
		// ContentPreviewURI is the URI of the digital art preview content (i.e. a thumbnail).
		ContentPreviewURI string
		// ContentMimetype is the content's MIME type (e.g. 'image/jpeg')
		ContentMimetype string
		// Edition number of the given NFT. Editions are unique for the same master,
		// identified by the asset ID.
		Edition uint64
		// MaxEdition is the number of editions that may have been produced
		// for the given master. This number can't be exceeded by the contract,
		// but not all the editions may have been minted (yet or ever).
		// If maxEdition == 1, the given NFT is one-of-a-kind.
		MaxEdition uint64
		// MetadataURI is a URI of the full digital art's metadata JSON
		// as it existed at the time the master was sealed.
		MetadataURI string
		// Record is the ChainLocker record ID of the full metadata JSON
		// as it existed at the time the master was sealed.
		Record string
		// AssetHead is the ChainLocker asset head ID of the full metadata JSON.
		// It can be used to retrieve the current metadata JSON (if changed).
		AssetHead string
	}

	// MintOnDemandParameters provides inputs for "digitalart_mint_on_demand_flow" and
	// "digitalart_mint_on_demand" transaction templates.
	// If Metadata is nil, the transactions won't include checks if the master is sealed
	// (and sealing it, if it's not).
	MintOnDemandParameters struct {
		Metadata *DigitalArtMetadata
		Profile  *evergreen.Profile
	}
)

func DigitalArtMetadataFromCadence(val cadence.Value) (*DigitalArtMetadata, error) {
	if opt, ok := val.(cadence.Optional); ok {
		if opt.Value == nil {
			return nil, nil
		}
		val = opt.Value
	}

	valStruct, ok := val.(cadence.Struct)
	if !ok || valStruct.StructType.QualifiedIdentifier != "DigitalArt.Metadata" || len(valStruct.FieldsMappedByName()) != 13 {
		return nil, errors.New("bad Metadata value")
	}

	allFields := valStruct.FieldsMappedByName()

	res := DigitalArtMetadata{
		Name:              string(allFields["name"].(cadence.String)),
		Artist:            string(allFields["artist"].(cadence.String)),
		Description:       string(allFields["description"].(cadence.String)),
		Type:              string(allFields["type"].(cadence.String)),
		ContentURI:        string(allFields["contentURI"].(cadence.String)),
		ContentPreviewURI: string(allFields["contentPreviewURI"].(cadence.String)),
		ContentMimetype:   string(allFields["mimetype"].(cadence.String)),
		Edition:           uint64(allFields["edition"].(cadence.UInt64)),
		MaxEdition:        uint64(allFields["maxEdition"].(cadence.UInt64)),
		Asset:             string(allFields["asset"].(cadence.String)),
		MetadataURI:       string(allFields["metadataURI"].(cadence.String)),
		Record:            string(allFields["record"].(cadence.String)),
		AssetHead:         string(allFields["assetHead"].(cadence.String)),
	}

	return &res, nil
}

func DigitalArtMetadataToCadence(metadata *DigitalArtMetadata, digitalArtAddr flow.Address) cadence.Value {
	return cadence.NewStruct([]cadence.Value{
		cadence.String(metadata.Name),
		cadence.String(metadata.Artist),
		cadence.String(metadata.Description),
		cadence.String(metadata.Type),
		cadence.String(metadata.ContentURI),
		cadence.String(metadata.ContentPreviewURI),
		cadence.String(metadata.ContentMimetype),
		cadence.UInt64(metadata.Edition),
		cadence.UInt64(metadata.MaxEdition),
		cadence.String(metadata.Asset),
		cadence.String(metadata.MetadataURI),
		cadence.String(metadata.Record),
		cadence.String(metadata.AssetHead),
	}).WithType(cadence.NewStructType(
		common.AddressLocation{
			Address: common.Address(digitalArtAddr),
			Name:    common.AddressLocationPrefix,
		},
		"DigitalArt.Metadata",
		metadataCadenceFields,
		nil,
	))
}

var metadataCadenceFields = []cadence.Field{
	{
		Identifier: "name",
		Type:       cadence.StringType,
	},
	{
		Identifier: "artist",
		Type:       cadence.StringType,
	},
	{
		Identifier: "description",
		Type:       cadence.StringType,
	},
	{
		Identifier: "type",
		Type:       cadence.StringType,
	},
	{
		Identifier: "contentURI",
		Type:       cadence.StringType,
	},
	{
		Identifier: "contentPreviewURI",
		Type:       cadence.StringType,
	},
	{
		Identifier: "mimetype",
		Type:       cadence.StringType,
	},
	{
		Identifier: "edition",
		Type:       cadence.UInt64Type,
	},
	{
		Identifier: "maxEdition",
		Type:       cadence.UInt64Type,
	},
	{
		Identifier: "asset",
		Type:       cadence.StringType,
	},
	{
		Identifier: "metadataURI",
		Type:       cadence.StringType,
	},
	{
		Identifier: "record",
		Type:       cadence.StringType,
	},
	{
		Identifier: "assetHead",
		Type:       cadence.StringType,
	},
}
