package evergreen

import (
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
	if !ok || valStruct.StructType.QualifiedIdentifier != "Evergreen.Role" || len(valStruct.Fields) != 6 {
		return nil, errors.New("bad Evergreen Role value")
	}

	var receiverPath string
	if opt := valStruct.Fields[5].(cadence.Optional).Value; opt != nil {
		receiverPath = opt.String()
	}

	res := Role{
		ID:                        valStruct.Fields[0].ToGoValue().(string),
		Description:               valStruct.Fields[1].ToGoValue().(string),
		InitialSaleCommission:     iinft.ToFloat64(valStruct.Fields[2]),
		SecondaryMarketCommission: iinft.ToFloat64(valStruct.Fields[3]),
		Address:                   flow.BytesToAddress(valStruct.Fields[4].(cadence.Address).Bytes()),
		ReceiverPath:              receiverPath,
	}

	return &res, nil
}

func RoleToCadence(role *Role, evergreenAddr flow.Address) (cadence.Value, error) {
	var receiverPath cadence.Value
	if role.ReceiverPath != "" {
		path, err := iinft.StringToPath(role.ReceiverPath)
		if err != nil {
			return nil, err
		}
		receiverPath = path
	}
	return cadence.NewStruct([]cadence.Value{
		cadence.String(role.ID),
		cadence.String(role.Description),
		iinft.UFix64FromFloat64(role.InitialSaleCommission),
		iinft.UFix64FromFloat64(role.SecondaryMarketCommission),
		cadence.BytesToAddress(role.Address.Bytes()),
		cadence.NewOptional(receiverPath),
	}).WithType(&cadence.StructType{
		Location: common.AddressLocation{
			Address: common.Address(evergreenAddr),
			Name:    common.AddressLocationPrefix,
		},
		QualifiedIdentifier: "Evergreen.Role",
		Fields:              roleCadenceFields,
	}), nil
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
		ID:          uint32(valStruct.Fields[0].(cadence.UInt32)),
		Description: valStruct.Fields[1].ToGoValue().(string),
		Roles:       []*Role{},
	}

	rolesArray, ok := valStruct.Fields[2].(cadence.Array)
	if !ok {
		return nil, errors.New("bad Evergreen Profile value")
	}
	for _, roleVal := range rolesArray.Values {
		role, err := RoleFromCadence(roleVal)
		if err != nil {
			return nil, err
		}
		res.Roles = append(res.Roles, role)
	}

	return &res, nil
}

func ProfileToCadence(profile *Profile, evergreenAddr flow.Address) (cadence.Value, error) {

	roles := make([]cadence.Value, len(profile.Roles))
	i := 0
	var err error
	for _, role := range profile.Roles {
		roles[i], err = RoleToCadence(role, evergreenAddr)
		if err != nil {
			return nil, err
		}
		i++
	}

	return cadence.NewStruct([]cadence.Value{
		cadence.UInt32(profile.ID),
		cadence.String(profile.Description),
		cadence.NewArray(roles),
	}).WithType(&cadence.StructType{
		Location: common.AddressLocation{
			Address: common.Address(evergreenAddr),
			Name:    common.AddressLocationPrefix,
		},
		QualifiedIdentifier: "Evergreen.Profile",
		Fields:              profileCadenceFields,
	}), nil
}

var roleCadenceFields = []cadence.Field{
	{
		Identifier: "id",
		Type:       cadence.StringType{},
	},
	{
		Identifier: "description",
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
	{
		Identifier: "receiverPath",
		Type:       cadence.OptionalType{},
	},
}

var profileCadenceFields = []cadence.Field{
	{
		Identifier: "id",
		Type:       cadence.StringType{},
	},
	{
		Identifier: "description",
		Type:       cadence.StringType{},
	},
	{
		Identifier: "roles",
		Type:       cadence.DictionaryType{},
	},
}
