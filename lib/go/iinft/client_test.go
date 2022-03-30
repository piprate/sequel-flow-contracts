package iinft

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp})
}

func TestNewGoWithTheFlowFS(t *testing.T) {
	client, err := NewGoWithTheFlowFS("../../..", "emulator", true)
	require.NoError(t, err)

	_, err = client.CreateAccountsE("emulator-account")
	require.NoError(t, err)

	client.InitializeContracts()
}

func TestNewGoWithTheFlowEmbedded(t *testing.T) {
	client, err := NewGoWithTheFlowEmbedded("emulator", true)
	require.NoError(t, err)

	_, err = client.CreateAccountsE("emulator-account")
	require.NoError(t, err)

	client.InitializeContracts()
}
