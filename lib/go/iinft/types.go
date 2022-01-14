package iinft

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flow-go-sdk"
)

type (
	Metadata struct {
		MetadataLink       string
		Name               string
		Artist             string
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
)

func MetadataFromCadence(val cadence.Value) (*Metadata, error) {
	if opt, ok := val.(cadence.Optional); ok {
		if opt.Value == nil {
			return nil, nil
		}
		val = opt.Value
	}

	valStruct, ok := val.(cadence.Struct)
	if !ok || valStruct.StructType.QualifiedIdentifier != "DigitalArt.Metadata" || len(valStruct.Fields) != 13 {
		return nil, errors.New("bad Metadata value")
	}

	res := Metadata{
		MetadataLink:       valStruct.Fields[0].ToGoValue().(string),
		Name:               valStruct.Fields[1].ToGoValue().(string),
		Artist:             valStruct.Fields[2].ToGoValue().(string),
		Description:        valStruct.Fields[3].ToGoValue().(string),
		Type:               valStruct.Fields[4].ToGoValue().(string),
		ContentLink:        valStruct.Fields[5].ToGoValue().(string),
		ContentPreviewLink: valStruct.Fields[6].ToGoValue().(string),
		Mimetype:           valStruct.Fields[7].ToGoValue().(string),
		Edition:            uint64(valStruct.Fields[8].(cadence.UInt64)),
		MaxEdition:         uint64(valStruct.Fields[9].(cadence.UInt64)),
		Asset:              valStruct.Fields[10].ToGoValue().(string),
		Record:             valStruct.Fields[11].ToGoValue().(string),
		AssetHead:          valStruct.Fields[12].ToGoValue().(string),
	}

	return &res, nil
}

func MetadataToCadence(metadata *Metadata, digitalArtAddr flow.Address) cadence.Value {
	return cadence.NewStruct([]cadence.Value{
		cadence.String(metadata.MetadataLink),
		cadence.String(metadata.Name),
		cadence.String(metadata.Artist),
		cadence.String(metadata.Description),
		cadence.String(metadata.Type),
		cadence.String(metadata.ContentLink),
		cadence.String(metadata.ContentPreviewLink),
		cadence.String(metadata.Mimetype),
		cadence.UInt64(metadata.Edition),
		cadence.UInt64(metadata.MaxEdition),
		cadence.String(metadata.Asset),
		cadence.String(metadata.Record),
		cadence.String(metadata.AssetHead),
	}).WithType(&cadence.StructType{
		Location: common.AddressLocation{
			Address: common.Address(digitalArtAddr),
			Name:    common.AddressLocationPrefix,
		},
		QualifiedIdentifier: "DigitalArt.Metadata",
		Fields:              metadataCadenceFields,
	})
}

func ToFloat64(value cadence.Value) float64 {
	val, _ := strconv.ParseFloat(value.(cadence.UFix64).String(), 64)
	return val
}

func UFix64FromFloat64(v float64) cadence.Value {
	cv, err := cadence.NewUFix64(fmt.Sprintf("%.4f", v))
	if err != nil {
		panic(err)
	}
	return cv
}

var metadataCadenceFields = []cadence.Field{
	{
		Identifier: "metadataLink",
		Type:       cadence.StringType{},
	},
	{
		Identifier: "name",
		Type:       cadence.StringType{},
	},
	{
		Identifier: "artist",
		Type:       cadence.StringType{},
	},
	{
		Identifier: "description",
		Type:       cadence.StringType{},
	},
	{
		Identifier: "type",
		Type:       cadence.StringType{},
	},
	{
		Identifier: "contentLink",
		Type:       cadence.StringType{},
	},
	{
		Identifier: "contentPreviewLink",
		Type:       cadence.StringType{},
	},
	{
		Identifier: "mimetype",
		Type:       cadence.StringType{},
	},
	{
		Identifier: "edition",
		Type:       cadence.UInt64Type{},
	},
	{
		Identifier: "maxEdition",
		Type:       cadence.UInt64Type{},
	},
	{
		Identifier: "asset",
		Type:       cadence.StringType{},
	},
	{
		Identifier: "record",
		Type:       cadence.StringType{},
	},
	{
		Identifier: "assetHead",
		Type:       cadence.StringType{},
	},
}
