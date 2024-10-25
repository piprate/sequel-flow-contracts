package iinft

import (
	"context"
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
	client, err := NewGoWithTheFlowFS("../../..", "emulator", true, false)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = client.CreateAccountsE(ctx, "emulator-account")
	require.NoError(t, err)

	err = client.InitializeContractsE(ctx)
	require.NoError(t, err)
}

func TestNewGoWithTheFlowEmbedded(t *testing.T) {
	client, err := NewGoWithTheFlowEmbedded("emulator", true, false)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = client.CreateAccountsE(ctx, "emulator-account")
	require.NoError(t, err)

	err = client.InitializeContractsE(ctx)
	require.NoError(t, err)
}

func TestNewGoWithTheFlowEmbedded_WithTxFees(t *testing.T) {
	client, err := NewGoWithTheFlowEmbedded("emulator", true, true)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = client.DoNotPrependNetworkToAccountNames().CreateAccountsE(ctx, "emulator-account")
	require.NoError(t, err)
}
