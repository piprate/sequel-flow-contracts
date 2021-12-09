package gwtf

import (
	"testing"

	"github.com/onflow/cadence"
	"github.com/stretchr/testify/assert"
)

func TestCadenceValueToJsonString(t *testing.T) {

	t.Parallel()
	t.Run("Empty value should be empty json object", func(t *testing.T) {
		value := CadenceValueToJsonString(nil)
		assert.Equal(t, "{}", value)
	})

	t.Run("Empty optional should be empty string", func(t *testing.T) {
		value := CadenceValueToJsonString(cadence.NewOptional(nil))
		assert.Equal(t, `""`, value)
	})
	t.Run("Unwrap optional", func(t *testing.T) {
		fooStr, _ := cadence.NewString("foo")
		value := CadenceValueToJsonString(cadence.NewOptional(fooStr))
		assert.Equal(t, `"foo"`, value)
	})
	t.Run("Array", func(t *testing.T) {
		fooStr, _ := cadence.NewString("foo")
		barStr, _ := cadence.NewString("bar")
		value := CadenceValueToJsonString(cadence.NewArray([]cadence.Value{fooStr, barStr}))
		assert.Equal(t, `[
    "foo",
    "bar"
]`, value)
	})

	t.Run("Dictionary", func(t *testing.T) {
		fooStr, _ := cadence.NewString("foo")
		barStr, _ := cadence.NewString("bar")
		dict := cadence.NewDictionary([]cadence.KeyValuePair{{Key: fooStr, Value: barStr}})
		value := CadenceValueToJsonString(dict)
		assert.Equal(t, `{
    "foo": "bar"
}`, value)
	})

	t.Run("Dictionary nested", func(t *testing.T) {
		fooStr, _ := cadence.NewString("foo")
		barStr, _ := cadence.NewString("bar")
		subdictStr, _ := cadence.NewString("subdict")
		subDict := cadence.NewDictionary([]cadence.KeyValuePair{{Key: fooStr, Value: barStr}})
		dict := cadence.NewDictionary([]cadence.KeyValuePair{{Key: fooStr, Value: barStr}, {Key: subdictStr, Value: subDict}})
		value := CadenceValueToJsonString(dict)
		assert.Equal(t, `{
    "foo": "bar",
    "subdict": {
        "foo": "bar"
    }
}`, value)
	})

	t.Run("Dictionary", func(t *testing.T) {
		dict := cadence.NewDictionary([]cadence.KeyValuePair{{Key: cadence.NewUInt64(1), Value: cadence.NewUInt64(1)}})
		value := CadenceValueToJsonString(dict)
		assert.Equal(t, `{
    "1": "1"
}`, value)
	})

	t.Run("Struct", func(t *testing.T) {
		barStr, _ := cadence.NewString("bar")
		s := cadence.Struct{
			Fields: []cadence.Value{barStr},
			StructType: &cadence.StructType{
				Fields: []cadence.Field{{
					Identifier: "foo",
					Type:       cadence.StringType{},
				}},
			},
		}
		value := CadenceValueToJsonString(s)
		assert.Equal(t, `{
    "foo": "bar"
}`, value)
	})

}
