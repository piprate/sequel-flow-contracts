package iinft

import (
	"errors"
	"strconv"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

const (
	EvergreenRoleArtist    = "Artist"
	EvergreenRolePlatform  = "Platform"
	EvergreenRoleCollector = "Collector"
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
		EvergreenProfile   *EvergreenProfile
	}

	EvergreenRole struct {
		Role                      string       `json:"role"`
		InitialSaleCommission     float64      `json:"initialSaleCommission,omitempty"`
		SecondaryMarketCommission float64      `json:"secondaryMarketCommission,omitempty"`
		Address                   flow.Address `json:"addr,omitempty"`
	}

	EvergreenProfile struct {
		ID    uint32                    `json:"id"`
		Roles map[string]*EvergreenRole `json:"roles"`
	}
)

func NewEvergreenRoleFromCadence(val cadence.Value) (*EvergreenRole, error) {
	if opt, ok := val.(cadence.Optional); ok {
		if opt.Value == nil {
			return nil, nil
		}
		val = opt.Value
	}

	valStruct, ok := val.(cadence.Struct)
	if !ok || valStruct.StructType.QualifiedIdentifier != "Evergreen.Role" || len(valStruct.Fields) != 4 {
		return nil, errors.New("bad Evergreen Role value")
	}

	res := EvergreenRole{
		Role:                      valStruct.Fields[0].ToGoValue().(string),
		InitialSaleCommission:     ToFloat64(valStruct.Fields[1]),
		SecondaryMarketCommission: ToFloat64(valStruct.Fields[2]),
		Address:                   flow.BytesToAddress(valStruct.Fields[3].(cadence.Address).Bytes()),
	}

	return &res, nil
}

func NewEvergreenProfileFromCadence(val cadence.Value) (*EvergreenProfile, error) {
	if opt, ok := val.(cadence.Optional); ok {
		if opt.Value == nil {
			return nil, nil
		}
		val = opt.Value
	}

	valStruct, ok := val.(cadence.Struct)
	if !ok || valStruct.StructType.QualifiedIdentifier != "Evergreen.Profile" || len(valStruct.Fields) != 3 {
		return nil, errors.New("bad Evergreen Profile value")
	}

	res := EvergreenProfile{
		ID:    uint32(valStruct.Fields[0].(cadence.UInt32)),
		Roles: map[string]*EvergreenRole{},
	}

	rolesDict := valStruct.Fields[1].(cadence.Dictionary)
	var err error
	for _, pair := range rolesDict.Pairs {
		res.Roles[pair.Key.ToGoValue().(string)], err = NewEvergreenRoleFromCadence(pair.Value)
		if err != nil {
			return nil, err
		}
	}

	return &res, nil
}

func NewMetadataFromCadence(val cadence.Value) (*Metadata, error) {
	if opt, ok := val.(cadence.Optional); ok {
		if opt.Value == nil {
			return nil, nil
		}
		val = opt.Value
	}

	valStruct, ok := val.(cadence.Struct)
	if !ok || valStruct.StructType.QualifiedIdentifier != "DigitalArt.Metadata" || len(valStruct.Fields) != 14 {
		return nil, errors.New("bad Metadata value")
	}

	profile, err := NewEvergreenProfileFromCadence(valStruct.Fields[13])
	if err != nil {
		return nil, err
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
		EvergreenProfile:   profile,
	}

	return &res, nil
}

func ToFloat64(value cadence.Value) float64 {
	val, _ := strconv.ParseFloat(value.(cadence.UFix64).String(), 64)
	return val
}
