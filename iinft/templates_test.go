package iinft_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	. "github.com/piprate/sequel-flow-contracts/iinft"
	"github.com/piprate/sequel-flow-contracts/iinft/evergreen"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp})
}

func TestNewEngine_emulator(t *testing.T) {
	client, err := NewInMemoryConnectorEmbedded(false)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = client.CreateAccountsE(ctx, "emulator-account")
	require.NoError(t, err)

	err = client.InitializeContractsE(ctx)
	require.NoError(t, err)

	_, err = NewTemplateEngine(client)
	require.NoError(t, err)
}

func TestNewEngine_emulatorWithFees(t *testing.T) {
	client, err := NewInMemoryConnectorEmbedded(true)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = client.DoNotPrependNetworkToAccountNames().CreateAccountsE(ctx, "emulator-account")
	require.NoError(t, err)

	te, err := NewTemplateEngine(client)
	require.NoError(t, err)

	adminAcct := client.Account("emulator-sequel-admin")
	_ = te.NewTransaction("account_fund_flow").
		Argument(cadence.NewAddress(adminAcct.Address)).
		UFix64Argument("1000.0").
		SignProposeAndPayAsService().
		Test(t).
		AssertSuccess()

	err = client.InitializeContractsE(ctx)
	require.NoError(t, err)
}

func TestNewEngine_testnet(t *testing.T) {
	client, err := NewNetworkConnectorEmbedded("testnet")
	require.NoError(t, err)

	_, err = NewTemplateEngine(client)
	require.NoError(t, err)
}

func TestNewEngine_mainnet(t *testing.T) {
	client, err := NewNetworkConnectorEmbedded("mainnet")
	require.NoError(t, err)

	_, err = NewTemplateEngine(client)
	require.NoError(t, err)
}

func TestEngine_GetStandardScript(t *testing.T) {
	client, err := NewNetworkConnectorEmbedded("testnet")
	require.NoError(t, err)

	ctx := context.Background()

	_, err = client.CreateAccountsE(ctx, "emulator-account")
	require.NoError(t, err)

	err = client.InitializeContractsE(ctx)
	require.NoError(t, err)

	e, err := NewTemplateEngine(client)
	require.NoError(t, err)

	res := e.GetStandardScript("catalog_get_collection_tokens")

	println(res)
}

func TestEngine_GetStandardScript_Versus(t *testing.T) {
	client, err := NewNetworkConnectorEmbedded("mainnet")
	require.NoError(t, err)

	ctx := context.Background()

	_, err = client.CreateAccountsE(ctx, "emulator-account")
	require.NoError(t, err)

	err = client.InitializeContractsE(ctx)
	require.NoError(t, err)

	e, err := NewTemplateEngine(client)
	require.NoError(t, err)

	res := e.GetStandardScript("versus_get_art")

	println(res)
}

func TestEngine_GetCustomScript_MOD_Flow(t *testing.T) {
	client, err := NewNetworkConnectorEmbedded("mainnet")
	require.NoError(t, err)

	ctx := context.Background()

	_, err = client.CreateAccountsE(ctx, "emulator-account")
	require.NoError(t, err)

	err = client.InitializeContractsE(ctx)
	require.NoError(t, err)

	e, err := NewTemplateEngine(client)
	require.NoError(t, err)

	res := e.GetCustomScript("digitalart_mint_on_demand_flow", &MintOnDemandParameters{
		Metadata: &DigitalArtMetadata{
			MetadataURI:       "ipfs://QmMetadata",
			Name:              "Pure Art",
			Artist:            "did:sequel:artist",
			Description:       "Digital art in its purest form",
			Type:              "Image",
			ContentURI:        "ipfs://QmContent",
			ContentPreviewURI: "ipfs://QmPreview",
			ContentMimetype:   "image/jpeg",
			MaxEdition:        4,
			Asset:             "did:sequel:asset-id",
			Record:            "record-id",
			AssetHead:         "asset-head-id",
		},
		Profile: &evergreen.Profile{
			ID: "did:sequel:evergreen1",
			Roles: []*evergreen.Role{
				{
					ID:                        evergreen.RoleArtist,
					InitialSaleCommission:     0.8,
					SecondaryMarketCommission: 0.2,
					Address:                   flow.HexToAddress("0xf669cb8d41ce0c74"),
				},
			},
		},
	})

	println(res)
}
