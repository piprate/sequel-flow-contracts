package scripts_test

import (
	"os"
	"testing"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/evergreen"
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

	_, err = client.CreateAccountsE("emulator-account")
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

	_, err = scripts.NewEngine(client, false)
	require.NoError(t, err)
}

func TestEngine_GetStandardScript(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowEmbedded("testnet", false)
	require.NoError(t, err)

	_, err = client.CreateAccountsE("emulator-account")
	require.NoError(t, err)

	client.InitializeContracts()

	e, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	res := e.GetStandardScript("xtingles_get_collection")

	println(res)
}

func TestEngine_GetStandardScript_Versus(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowEmbedded("mainnet", false)
	require.NoError(t, err)

	_, err = client.CreateAccountsE("emulator-account")
	require.NoError(t, err)

	client.InitializeContracts()

	e, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	res := e.GetStandardScript("versus_get_art")

	println(res)
}

func TestEngine_GetCustomScript_MOD_FUSD(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowEmbedded("mainnet", false)
	require.NoError(t, err)

	_, err = client.CreateAccountsE("emulator-account")
	require.NoError(t, err)

	client.InitializeContracts()

	e, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	res := e.GetCustomScript("digitalart_mint_on_demand_fusd", &scripts.MindOnDemandParameters{
		Metadata: &iinft.Metadata{
			MetadataLink:       "QmMetadata",
			Name:               "Pure Art",
			Artist:             "did:sequel:artist",
			Description:        "Digital art in its purest form",
			Type:               "Image",
			ContentLink:        "QmContent",
			ContentPreviewLink: "QmPreview",
			Mimetype:           "image/jpeg",
			MaxEdition:         4,
			Asset:              "did:sequel:asset-id",
			Record:             "record-id",
			AssetHead:          "asset-head-id",
		},
		Profile: &evergreen.Profile{
			ID: 0,
			Roles: []*evergreen.Role{
				{
					Role:                      evergreen.RoleArtist,
					InitialSaleCommission:     0.8,
					SecondaryMarketCommission: 0.2,
					Address:                   flow.HexToAddress("0xf669cb8d41ce0c74"),
				},
			},
		},
	})

	println(res)
}
