package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewGoWithTheFlowEmbedded(t *testing.T) {
	client, err := NewGoWithTheFlowEmbedded("emulator", true)
	require.NoError(t, err)

	client.InitializeContracts()
}
