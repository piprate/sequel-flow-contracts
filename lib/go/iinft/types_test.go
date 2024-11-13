package iinft_test

import (
	"testing"

	. "github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/stretchr/testify/assert"
)

func TestUFix64ToString(t *testing.T) {
	assert.Equal(t, "200.0", UFix64ToString(200.0))
	assert.Equal(t, "200.25", UFix64ToString(200.25))
	assert.Equal(t, "200.25252525", UFix64ToString(200.25252525))
}

func TestStringToPath(t *testing.T) {
	val, err := StringToPath("/private/test")
	assert.NoError(t, err)
	assert.Equal(t, "private", val.Domain.Identifier())
	assert.Equal(t, "test", val.Identifier)

	_, err = StringToPath("/bad/test")
	assert.Error(t, err)

	_, err = StringToPath("super/bad")
	assert.Error(t, err)

	_, err = StringToPath("/storage/bad/test")
	assert.Error(t, err)
}
