package iinft

import (
	"errors"
	"strconv"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
)

const (
	ParticipationRoleArtist    = "Artist"
	ParticipationRolePlatform  = "Platform"
	ParticipationRoleCollector = "Collector"
)

type (
	Metadata struct {
		MetadataLink         string
		Name                 string
		Artist               string
		Description          string
		Type                 string
		ContentLink          string
		ContentPreviewLink   string
		Mimetype             string
		Edition              uint64
		MaxEdition           uint64
		Asset                string
		Record               string
		AssetHead            string
		ParticipationProfile *ParticipationProfile
	}

	ParticipationRole struct {
		Role                      string       `json:"role"`
		InitialSaleCommission     float64      `json:"initialSaleCommission,omitempty"`
		SecondaryMarketCommission float64      `json:"secondaryMarketCommission,omitempty"`
		Address                   flow.Address `json:"addr,omitempty"`
	}

	ParticipationProfile struct {
		ID    uint32                        `json:"id"`
		Roles map[string]*ParticipationRole `json:"roles"`
	}
)

func NewParticipationRoleFromCadence(val cadence.Value) (*ParticipationRole, error) {
	if opt, ok := val.(cadence.Optional); ok {
		if opt.Value == nil {
			return nil, nil
		}
		val = opt.Value
	}

	valStruct, ok := val.(cadence.Struct)
	if !ok || valStruct.StructType.QualifiedIdentifier != "Participation.Role" || len(valStruct.Fields) != 4 {
		return nil, errors.New("bad Participation Role value")
	}

	res := ParticipationRole{
		Role:                      valStruct.Fields[0].ToGoValue().(string),
		InitialSaleCommission:     ToFloat64(valStruct.Fields[1]),
		SecondaryMarketCommission: ToFloat64(valStruct.Fields[2]),
		Address:                   flow.BytesToAddress(valStruct.Fields[3].(cadence.Address).Bytes()),
	}

	return &res, nil
}

func NewParticipationProfileFromCadence(val cadence.Value) (*ParticipationProfile, error) {
	if opt, ok := val.(cadence.Optional); ok {
		if opt.Value == nil {
			return nil, nil
		}
		val = opt.Value
	}

	valStruct, ok := val.(cadence.Struct)
	if !ok || valStruct.StructType.QualifiedIdentifier != "Participation.Profile" || len(valStruct.Fields) != 3 {
		return nil, errors.New("bad Participation Profile value")
	}

	res := ParticipationProfile{
		ID:    uint32(valStruct.Fields[0].(cadence.UInt32)),
		Roles: map[string]*ParticipationRole{},
	}

	rolesDict := valStruct.Fields[1].(cadence.Dictionary)
	var err error
	for _, pair := range rolesDict.Pairs {
		res.Roles[pair.Key.ToGoValue().(string)], err = NewParticipationRoleFromCadence(pair.Value)
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

	profile, err := NewParticipationProfileFromCadence(valStruct.Fields[13])
	if err != nil {
		return nil, err
	}

	res := Metadata{
		MetadataLink:         valStruct.Fields[0].ToGoValue().(string),
		Name:                 valStruct.Fields[1].ToGoValue().(string),
		Artist:               valStruct.Fields[2].ToGoValue().(string),
		Description:          valStruct.Fields[3].ToGoValue().(string),
		Type:                 valStruct.Fields[4].ToGoValue().(string),
		ContentLink:          valStruct.Fields[5].ToGoValue().(string),
		ContentPreviewLink:   valStruct.Fields[6].ToGoValue().(string),
		Mimetype:             valStruct.Fields[7].ToGoValue().(string),
		Edition:              uint64(valStruct.Fields[8].(cadence.UInt64)),
		MaxEdition:           uint64(valStruct.Fields[9].(cadence.UInt64)),
		Asset:                valStruct.Fields[10].ToGoValue().(string),
		Record:               valStruct.Fields[11].ToGoValue().(string),
		AssetHead:            valStruct.Fields[12].ToGoValue().(string),
		ParticipationProfile: profile,
	}

	return &res, nil
}

func ToFloat64(value cadence.Value) float64 {
	val, _ := strconv.ParseFloat(value.(cadence.UFix64).String(), 64)
	return val
}
