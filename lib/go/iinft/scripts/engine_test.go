package scripts_test

import (
	"os"
	"testing"
	"time"

	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/scripts"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp})
}

func TestNewEngine_emulator(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowEmbedded("emulator", true)
	require.NoError(t, err)

	client.InitializeContracts()

	_, err = scripts.NewEngine(client, false)
	require.NoError(t, err)
}

func TestNewEngine_testnet(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowEmbedded("testnet", false)
	require.NoError(t, err)

	_, err = scripts.NewEngine(client, false)
	require.NoError(t, err)
}

func TestNewEngine_mainnet(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowEmbedded("mainnet", false)
	require.NoError(t, err)

	// this call should fail because we don't have DigitalArt deployed on mainnet yet
	_, err = scripts.NewEngine(client, false)
	require.Error(t, err)
}
