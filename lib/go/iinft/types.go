package iinft

import (
	"errors"

	"github.com/onflow/cadence"
	sdk "github.com/onflow/flow-go-sdk"
)

type Metadata struct {
	MetadataLink       string
	Name               string
	Artist             string
	ArtistAddress      sdk.Address
	Description        string
	Type               string
	ContentLink        string
	ContentPreviewLink string
	Mimetype           string
	Edition            uint64
	MaxEdition         uint64
	Asset              string
	Record             string
	AssetHead          string
}

func ReadMetadata(val cadence.Value) (*Metadata, error) {
	if opt, ok := val.(cadence.Optional); ok {
		if opt.Value == nil {
			return nil, nil
		}
		val = opt.Value
	}

	valStruct, ok := val.(cadence.Struct)
	if !ok || valStruct.StructType.QualifiedIdentifier != "DigitalArt.Metadata" || len(valStruct.Fields) != 14 {
		println(valStruct.StructType.QualifiedIdentifier)
		println(len(valStruct.Fields))
		return nil, errors.New("bad Metadata value")
	}

	res := Metadata{
		MetadataLink:       valStruct.Fields[0].ToGoValue().(string),
		Name:               valStruct.Fields[1].ToGoValue().(string),
		Artist:             valStruct.Fields[2].ToGoValue().(string),
		ArtistAddress:      sdk.BytesToAddress(valStruct.Fields[3].(cadence.Address).Bytes()),
		Description:        valStruct.Fields[4].ToGoValue().(string),
		Type:               valStruct.Fields[5].ToGoValue().(string),
		ContentLink:        valStruct.Fields[6].ToGoValue().(string),
		ContentPreviewLink: valStruct.Fields[7].ToGoValue().(string),
		Mimetype:           valStruct.Fields[8].ToGoValue().(string),
		Edition:            uint64(valStruct.Fields[9].(cadence.UInt64)),
		MaxEdition:         uint64(valStruct.Fields[10].(cadence.UInt64)),
		Asset:              valStruct.Fields[11].ToGoValue().(string),
		Record:             valStruct.Fields[12].ToGoValue().(string),
		AssetHead:          valStruct.Fields[13].ToGoValue().(string),
	}

	return &res, nil
}
