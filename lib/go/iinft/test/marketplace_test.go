package test

import (
	"testing"

	"github.com/onflow/cadence"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/gwtf"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/scripts"
	"github.com/stretchr/testify/require"
)

func TestMarketplace_ListAndBuyWithFlow(t *testing.T) {
	client, err := iinft.NewGoWithTheFlowFS("../../../..", "emulator", true)
	require.NoError(t, err)

	client.InitializeContracts().CreateAccounts("emulator-account")

	se, err := scripts.NewEngine(client, false)
	require.NoError(t, err)

	// set up seller account

	sellerAccount := "user1"
	sellerAcct, err := client.State.Accounts().ByName("emulator-" + sellerAccount)
	require.NoError(t, err)

	_ = se.NewTransaction("account_setup").SignProposeAndPayAs(sellerAccount).Test(t).AssertSuccess()
	_ = se.NewTransaction("account_setup_flow_token").SignProposeAndPayAs(sellerAccount).Test(t).AssertSuccess()

	// set up buyer account

	buyerAccount := "user2"
	buyerAcct, err := client.State.Accounts().ByName("emulator-" + buyerAccount)
	require.NoError(t, err)

	_ = se.NewTransaction("account_setup").SignProposeAndPayAs(buyerAccount).Test(t).AssertSuccess()
	_ = se.NewTransaction("account_setup_flow_token").SignProposeAndPayAs(buyerAccount).Test(t).AssertSuccess()
	fundAccount(t, se, buyerAcct.Address(), "1000.0")

	metadata := &iinft.Metadata{
		MetadataLink:       "QmMetadata",
		Name:               "Pure Art",
		Artist:             "Arty",
		Description:        "Digital art in its purest form",
		Type:               "Image",
		ContentLink:        "QmContent",
		ContentPreviewLink: "QmPreview",
		Mimetype:           "image/jpeg",
		MaxEdition:         4,
		Asset:              "did:sequel:asset-id",
		Record:             "record-id",
		AssetHead:          "asset-head-id",
		ParticipationProfile: &iinft.ParticipationProfile{
			ID: 0,
			Roles: map[string]*iinft.ParticipationRole{
				iinft.ParticipationRoleArtist: {
					Role:                      iinft.ParticipationRoleArtist,
					InitialSaleCommission:     80.0,
					SecondaryMarketCommission: 20.0,
					Address:                   sellerAcct.Address(),
				},
			},
		},
	}

	_ = scripts.CreateMintSingleDigitalArtTx(se.GetStandardScript("digitalart_mint_single"), client, metadata, sellerAcct.Address()).
		SignProposeAndPayAs(sequelAccount).
		Test(t).
		AssertSuccess().
		AssertEmitEvent(gwtf.NewTestEvent("A.f8d6e0586b0a20c7.DigitalArt.Minted", map[string]interface{}{
			"id":      "0",
			"asset":   "did:sequel:asset-id",
			"edition": "1",
		}))

	var nftID uint64 = 0

	// Assert that the account's collection is correct
	checkTokenInDigitalArtCollection(t, se, sellerAcct.Address().String(), nftID)
	checkDigitalArtCollectionLen(t, se, sellerAcct.Address().String(), 1)
	checkDigitalArtCollectionLen(t, se, buyerAcct.Address().String(), 0)

	t.Run("Should be able to list an NFT in seller's Storefront", func(t *testing.T) {
		_ = se.NewTransaction("marketplace_add").
			SignProposeAndPayAs(sellerAccount).
			UInt64Argument(nftID).
			UFix64Argument("200.0").
			BooleanArgument(true).
			Test(t).
			AssertSuccess().
			AssertEmitEvent(gwtf.NewTestEvent(
				"A.f8d6e0586b0a20c7.NFTStorefront.ListingAvailable",
				map[string]interface{}{
					"ftVaultType":       "Type\u003cA.0ae53cb6e3f42a79.FlowToken.Vault\u003e()",
					"listingResourceID": "44",
					"nftID":             "0",
					"nftType":           "Type\u003cA.f8d6e0586b0a20c7.DigitalArt.NFT\u003e()",
					"price":             "200.00000000",
					"storefrontAddress": "0x1cf0e2f2f715450", // seller's address without leading 0
				}))
	})

	t.Run("Should be able to buy an NFT from seller's Storefront", func(t *testing.T) {
		_ = se.NewTransaction("marketplace_buy").
			SignProposeAndPayAs(buyerAccount).
			UInt64Argument(44).
			Argument(cadence.NewAddress(sellerAcct.Address())).
			Test(t).
			AssertSuccess()

		// Assert that the account's collection is correct
		checkTokenInDigitalArtCollection(t, se, buyerAcct.Address().String(), 0)
		checkDigitalArtCollectionLen(t, se, buyerAcct.Address().String(), 1)
		checkDigitalArtCollectionLen(t, se, sellerAcct.Address().String(), 0)
	})
}
