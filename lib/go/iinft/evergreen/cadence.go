package evergreen

import (
	"encoding/json"
	"errors"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flow-go-sdk"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
)

func RoleFromCadence(val cadence.Value) (*Role, error) {
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

	res := Role{
		Role:                      valStruct.Fields[0].ToGoValue().(string),
		InitialSaleCommission:     iinft.ToFloat64(valStruct.Fields[1]),
		SecondaryMarketCommission: iinft.ToFloat64(valStruct.Fields[2]),
		Address:                   flow.BytesToAddress(valStruct.Fields[3].(cadence.Address).Bytes()),
	}

	return &res, nil
}

func RoleToCadence(role *Role, evergreenAddr flow.Address) cadence.Value {
	return cadence.NewStruct([]cadence.Value{
		cadence.String(role.Role),
		iinft.UFix64FromFloat64(role.InitialSaleCommission),
		iinft.UFix64FromFloat64(role.SecondaryMarketCommission),
		cadence.BytesToAddress(role.Address.Bytes()),
	}).WithType(&cadence.StructType{
		Location: common.AddressLocation{
			Address: common.Address(evergreenAddr),
			Name:    common.AddressLocationPrefix,
		},
		QualifiedIdentifier: "Evergreen.Role",
		Fields:              roleCadenceFields,
	})
}

func ProfileFromCadence(val cadence.Value) (*Profile, error) {
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

	res := Profile{
		ID:    uint32(valStruct.Fields[0].(cadence.UInt32)),
		Roles: map[string]*Role{},
	}

	rolesDict := valStruct.Fields[1].(cadence.Dictionary)
	var err error
	for _, pair := range rolesDict.Pairs {
		//fmt.Printf("%+v\n", pair.Value)
		v, _ := json.MarshalIndent(pair.Value, "", "  ")
		println(string(v))
		res.Roles[pair.Key.ToGoValue().(string)], err = RoleFromCadence(pair.Value)
		if err != nil {
			return nil, err
		}
	}

	return &res, nil
}

func ProfileToCadence(profile *Profile, evergreenAddr flow.Address) cadence.Value {

	rolePairs := make([]cadence.KeyValuePair, len(profile.Roles))
	i := 0
	for _, role := range profile.Roles {
		rolePairs[i].Key = cadence.String(role.Role)
		rolePairs[i].Value = RoleToCadence(role, evergreenAddr)
		i++
	}

	return cadence.NewStruct([]cadence.Value{
		cadence.UInt32(profile.ID),
		cadence.NewDictionary(rolePairs),
	}).WithType(&cadence.StructType{
		Location: common.AddressLocation{
			Address: common.Address(evergreenAddr),
			Name:    common.AddressLocationPrefix,
		},
		QualifiedIdentifier: "Evergreen.Profile",
		Fields:              profileCadenceFields,
	})
}

var roleCadenceFields = []cadence.Field{
	{
		Identifier: "id",
		Type:       cadence.StringType{},
	},
	{
		Identifier: "initialSaleCommission",
		Type:       cadence.UFix64Type{},
	},
	{
		Identifier: "secondaryMarketCommission",
		Type:       cadence.UFix64Type{},
	},
	{
		Identifier: "address",
		Type:       cadence.AddressType{},
	},
}

var profileCadenceFields = []cadence.Field{
	{
		Identifier: "id",
		Type:       cadence.StringType{},
	},
	{
		Identifier: "roles",
		Type:       cadence.DictionaryType{},
	},
}
