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
