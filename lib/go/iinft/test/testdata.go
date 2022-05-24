package test

import (
	"github.com/onflow/flow-go-sdk"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/evergreen"
)

func SampleMetadata(maxEdition uint64) *iinft.DigitalArtMetadata {
	return &iinft.DigitalArtMetadata{
		MetadataURI:       "ipfs://QmMetadata",
		Name:              "Pure Art",
		Artist:            "did:sequel:artist",
		Description:       "Digital art in its purest form",
		Type:              "Image",
		ContentURI:        "ipfs://QmContent",
		ContentPreviewURI: "ipfs://QmPreview",
		ContentMimetype:   "image/jpeg",
		MaxEdition:        maxEdition,
		Asset:             "did:sequel:asset-id",
		Record:            "record-id",
		AssetHead:         "asset-head-id",
	}
}

func BasicEvergreenProfile(artist flow.Address) *evergreen.Profile {
	return &evergreen.Profile{
		ID: 1,
		Roles: []*evergreen.Role{
			{
				ID:                        evergreen.RoleArtist,
				InitialSaleCommission:     1.0,
				SecondaryMarketCommission: 0.05,
				Address:                   artist,
			},
		},
	}
}

func PrimaryOnlyEvergreenProfile(artist, platform flow.Address) *evergreen.Profile {
	return &evergreen.Profile{
		ID: 2,
		Roles: []*evergreen.Role{
			{
				ID:                        evergreen.RoleArtist,
				InitialSaleCommission:     0.8,
				SecondaryMarketCommission: 0.05,
				Address:                   artist,
			},
			{
				ID:                        evergreen.RolePlatform,
				InitialSaleCommission:     0.2,
				SecondaryMarketCommission: 0.0,
				Address:                   platform,
			},
		},
	}
}
