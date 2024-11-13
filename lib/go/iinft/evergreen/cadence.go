package evergreen

import (
	"errors"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/common"
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
	if !ok || valStruct.StructType.QualifiedIdentifier != "Evergreen.Role" || len(valStruct.FieldsMappedByName()) != 6 {
		return nil, errors.New("bad Evergreen Role value")
	}

	fieldMap := valStruct.FieldsMappedByName()

	var receiverPath string
	if opt := fieldMap["receiverPath"].(cadence.Optional).Value; opt != nil {
		receiverPath = opt.String()
	}

	res := Role{
		ID:                        string(fieldMap["id"].(cadence.String)),
		Description:               string(fieldMap["description"].(cadence.String)),
		InitialSaleCommission:     iinft.ToFloat64(fieldMap["initialSaleCommission"]),
		SecondaryMarketCommission: iinft.ToFloat64(fieldMap["secondaryMarketCommission"]),
		Address:                   flow.BytesToAddress(fieldMap["address"].(cadence.Address).Bytes()),
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
	}).WithType(cadence.NewStructType(
		common.AddressLocation{
			Address: common.Address(evergreenAddr),
			Name:    common.AddressLocationPrefix,
		},
		"Evergreen.Role",
		roleCadenceFields,
		nil,
	)), nil
}

func ProfileFromCadence(val cadence.Value) (*Profile, error) {
	if opt, ok := val.(cadence.Optional); ok {
		if opt.Value == nil {
			return nil, nil
		}
		val = opt.Value
	}

	valStruct, ok := val.(cadence.Struct)
	if !ok || valStruct.StructType.QualifiedIdentifier != "Evergreen.Profile" || len(valStruct.FieldsMappedByName()) != 3 {
		return nil, errors.New("bad Evergreen Profile value")
	}

	fieldMap := valStruct.FieldsMappedByName()

	res := Profile{
		ID:          string(fieldMap["id"].(cadence.String)),
		Description: string(fieldMap["description"].(cadence.String)),
		Roles:       []*Role{},
	}

	rolesArray, ok := fieldMap["roles"].(cadence.Array)
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
		cadence.String(profile.ID),
		cadence.String(profile.Description),
		cadence.NewArray(roles),
	}).WithType(cadence.NewStructType(
		common.AddressLocation{
			Address: common.Address(evergreenAddr),
			Name:    common.AddressLocationPrefix,
		},
		"Evergreen.Profile",
		profileCadenceFields(evergreenAddr),
		nil,
	)), nil
}

var roleCadenceFields = []cadence.Field{
	{
		Identifier: "id",
		Type:       cadence.StringType,
	},
	{
		Identifier: "description",
		Type:       cadence.StringType,
	},
	{
		Identifier: "initialSaleCommission",
		Type:       cadence.UFix64Type,
	},
	{
		Identifier: "secondaryMarketCommission",
		Type:       cadence.UFix64Type,
	},
	{
		Identifier: "address",
		Type:       cadence.AddressType,
	},
	{
		Identifier: "receiverPath",
		Type:       cadence.NewOptionalType(cadence.StringType),
	},
}

func profileCadenceFields(evergreenAddr flow.Address) []cadence.Field {
	return []cadence.Field{
		{
			Identifier: "id",
			Type:       cadence.StringType,
		},
		{
			Identifier: "description",
			Type:       cadence.StringType,
		},
		{
			Identifier: "roles",
			Type: cadence.NewDictionaryType(cadence.StringType, cadence.NewStructType(
				common.AddressLocation{
					Address: common.Address(evergreenAddr),
					Name:    common.AddressLocationPrefix,
				},
				"Evergreen.Role",
				roleCadenceFields,
				nil)),
		},
	}
}
