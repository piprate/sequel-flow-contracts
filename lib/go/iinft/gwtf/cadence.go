package gwtf

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/onflow/cadence"
)

// CadenceValueToJSONString converts a cadence.Value into a json pretty printed string
func CadenceValueToJSONString(value cadence.Value) string {
	if value == nil {
		return "{}"
	}

	result := CadenceValueToInterface(value)
	j, err := json.MarshalIndent(result, "", "    ")

	if err != nil {
		log.Fatal(err)
	}

	return string(j)
}

// CadenceValueToInterface convert a cadence.Value into interface{}
func CadenceValueToInterface(field cadence.Value) interface{} {
	if field == nil {
		return ""
	}

	switch typedField := field.(type) {
	case cadence.Optional:
		return CadenceValueToInterface(typedField.Value)
	case cadence.Dictionary:
		result := map[string]interface{}{}
		for _, item := range typedField.Pairs {
			key, err := strconv.Unquote(item.Key.String())
			if err != nil {
				result[item.Key.String()] = CadenceValueToInterface(item.Value)
				continue
			}

			result[key] = CadenceValueToInterface(item.Value)
		}
		return result
	case cadence.Struct:
		result := map[string]interface{}{}
		for name, subField := range typedField.FieldsMappedByName() {
			result[name] = CadenceValueToInterface(subField)
		}
		return result
	case cadence.Array:
		var result []interface{}
		for _, item := range typedField.Values {
			result = append(result, CadenceValueToInterface(item))
		}
		return result
	default:
		result, err := strconv.Unquote(field.String())
		if err != nil {
			return field.String()
		}
		return result
	}
}
