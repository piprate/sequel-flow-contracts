package test

import (
	"testing"

	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/test"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNFTDeployment(t *testing.T) {
	b := newBlockchain()

	_ = deployNFTContracts(t, b)
}

func TestSealDigitalArt(t *testing.T) {
	b := newBlockchain()

	contractsObj := deployNFTContracts(t, b)

	userAddress, userSigner := createAccount(t, b)
	setupAccount(t, b, userAddress, userSigner, contractsObj)

	sampleMetadata := &iinft.Metadata{
		MetadataLink:       "QmMetadata",
		Name:               "Pure Art",
		Artist:             "Arty",
		ArtistAddress:      userAddress,
		Description:        "Digital art in its purest form",
		Type:               "Image",
		ContentLink:        "QmContent",
		ContentPreviewLink: "QmPreview",
		Mimetype:           "image/jpeg",
		MaxEdition:         4,
		Asset:              "did:sequel:asset-id",
		Record:             "record-id",
		AssetHead:          "asset-head-id",
	}

	t.Run("Should be able to seal new digital art master", func(t *testing.T) {

		script, err := templates.GenerateSealDigitalArtScript(contractsObj.NFTAddress, contractsObj.DigitalArtAddress,
			sampleMetadata)
		require.NoError(t, err)

		tx := createTxWithTemplateAndAuthorizer(b, script, contractsObj.DigitalArtAddress)

		signAndSubmit(
			t, b, tx,
			[]flow.Address{
				b.ServiceKey().Address,
				contractsObj.DigitalArtAddress,
			},
			[]crypto.Signer{
				b.ServiceKey().Signer(),
				contractsObj.DigitalArtSigner,
			},
			false,
		)
	})

	t.Run("Shouldn't be able to seal the same digital art master twice", func(t *testing.T) {

		sampleMetadata2 := *sampleMetadata
		sampleMetadata2.Asset = "did:sequel:asset-2"

		// Seal the master
		script, err := templates.GenerateSealDigitalArtScript(contractsObj.NFTAddress, contractsObj.DigitalArtAddress,
			&sampleMetadata2)
		require.NoError(t, err)

		tx := createTxWithTemplateAndAuthorizer(b, script, contractsObj.DigitalArtAddress)

		signAndSubmit(
			t, b, tx,
			[]flow.Address{
				b.ServiceKey().Address,
				contractsObj.DigitalArtAddress,
			},
			[]crypto.Signer{
				b.ServiceKey().Signer(),
				contractsObj.DigitalArtSigner,
			},
			false,
		)

		// try again
		script, err = templates.GenerateSealDigitalArtScript(contractsObj.NFTAddress, contractsObj.DigitalArtAddress,
			&sampleMetadata2)
		require.NoError(t, err)

		tx = createTxWithTemplateAndAuthorizer(b, script, contractsObj.DigitalArtAddress)

		signAndSubmit(
			t, b, tx,
			[]flow.Address{
				b.ServiceKey().Address,
				contractsObj.DigitalArtAddress,
			},
			[]crypto.Signer{
				b.ServiceKey().Signer(),
				contractsObj.DigitalArtSigner,
			},
			true,
		)
	})
}

func TestCreateDigitalArt(t *testing.T) {
	b := newBlockchain()

	contractsObj := deployNFTContracts(t, b)

	userAddress, userSigner := createAccount(t, b)
	setupAccount(t, b, userAddress, userSigner, contractsObj)

	script := templates.GenerateInspectNFTSupplyScript(contractsObj.NFTAddress, contractsObj.DigitalArtAddress, "DigitalArt", 0)
	executeScriptAndCheck(t, b, script, nil)

	script = templates.GenerateInspectCollectionLenScript(
		contractsObj.NFTAddress,
		contractsObj.DigitalArtAddress,
		userAddress,
		"DigitalArt",
		"DigitalArt.CollectionPublicPath",
		0,
	)
	executeScriptAndCheck(t, b, script, nil)

	metadata := &iinft.Metadata{
		MetadataLink:       "QmMetadata",
		Name:               "Pure Art",
		Artist:             "Arty",
		ArtistAddress:      userAddress,
		Description:        "Digital art in its purest form",
		Type:               "Image",
		ContentLink:        "QmContent",
		ContentPreviewLink: "QmPreview",
		Mimetype:           "image/jpeg",
		MaxEdition:         4,
		Asset:              "did:sequel:asset-id",
		Record:             "record-id",
		AssetHead:          "asset-head-id",
	}

	script, err := templates.GenerateSealDigitalArtScript(contractsObj.NFTAddress, contractsObj.DigitalArtAddress,
		metadata)
	require.NoError(t, err)

	tx := createTxWithTemplateAndAuthorizer(b, script, contractsObj.DigitalArtAddress)

	signAndSubmit(
		t, b, tx,
		[]flow.Address{
			b.ServiceKey().Address,
			contractsObj.DigitalArtAddress,
		},
		[]crypto.Signer{
			b.ServiceKey().Signer(),
			contractsObj.DigitalArtSigner,
		},
		false,
	)

	t.Run("Should be able to mint a token", func(t *testing.T) {

		script, err = templates.GenerateMintNFTScript(metadata.Asset, contractsObj.NFTAddress, contractsObj.DigitalArtAddress,
			"DigitalArt", "DigitalArt.AdminStoragePath",
			"DigitalArt.CollectionPublicPath", userAddress)
		require.NoError(t, err)

		tx = createTxWithTemplateAndAuthorizer(b, script, contractsObj.DigitalArtAddress)

		signAndSubmit(
			t, b, tx,
			[]flow.Address{
				b.ServiceKey().Address,
				contractsObj.DigitalArtAddress,
			},
			[]crypto.Signer{
				b.ServiceKey().Signer(),
				contractsObj.DigitalArtSigner,
			},
			false,
		)

		// Assert that the account's collection is correct
		script = templates.GenerateInspectCollectionScript(
			contractsObj.NFTAddress,
			contractsObj.DigitalArtAddress,
			userAddress,
			"DigitalArt",
			"DigitalArt.CollectionPublicPath",
			0,
		)
		executeScriptAndCheck(t, b, script, nil)

		script = templates.GenerateInspectCollectionLenScript(
			contractsObj.NFTAddress,
			contractsObj.DigitalArtAddress,
			userAddress,
			"DigitalArt",
			"DigitalArt.CollectionPublicPath",
			1,
		)
		executeScriptAndCheck(t, b, script, nil)

		script = templates.GenerateInspectNFTSupplyScript(contractsObj.NFTAddress, contractsObj.DigitalArtAddress, "DigitalArt", 1)
		executeScriptAndCheck(t, b, script, nil)

		script, err = templates.GenerateGetMetadataScript(contractsObj.NFTAddress, contractsObj.DigitalArtAddress, "DigitalArt")
		require.NoError(t, err)

		arg1, err := jsoncdc.Encode(cadence.NewAddress(userAddress))
		require.NoError(t, err)

		arg2, err := jsoncdc.Encode(cadence.NewUInt64(0))
		require.NoError(t, err)

		val := executeScriptAndCheck(t, b, script, [][]byte{arg1, arg2})

		meta, err := iinft.ReadMetadata(val)
		require.NoError(t, err)

		assert.Equal(t, uint64(1), meta.Edition)

		script, err = templates.GenerateMintNFTScript(metadata.Asset, contractsObj.NFTAddress, contractsObj.DigitalArtAddress,
			"DigitalArt", "DigitalArt.AdminStoragePath",
			"DigitalArt.CollectionPublicPath", userAddress)
		require.NoError(t, err)

		tx = createTxWithTemplateAndAuthorizer(b, script, contractsObj.DigitalArtAddress)

		signAndSubmit(
			t, b, tx,
			[]flow.Address{
				b.ServiceKey().Address,
				contractsObj.DigitalArtAddress,
			},
			[]crypto.Signer{
				b.ServiceKey().Signer(),
				contractsObj.DigitalArtSigner,
			},
			false,
		)

		script, err = templates.GenerateGetMetadataScript(contractsObj.NFTAddress, contractsObj.DigitalArtAddress, "DigitalArt")
		require.NoError(t, err)

		arg1, err = jsoncdc.Encode(cadence.NewAddress(userAddress))
		require.NoError(t, err)

		arg2, err = jsoncdc.Encode(cadence.NewUInt64(1))
		require.NoError(t, err)

		val = executeScriptAndCheck(t, b, script, [][]byte{arg1, arg2})

		meta, err = iinft.ReadMetadata(val)
		require.NoError(t, err)

		assert.Equal(t, uint64(2), meta.Edition)
	})

	t.Run("Shouldn't be able to borrow a reference to an NFT that doesn't exist", func(t *testing.T) {

		// Assert that the account's collection is correct
		script := templates.GenerateInspectCollectionScript(
			contractsObj.NFTAddress,
			contractsObj.DigitalArtAddress,
			userAddress,
			"DigitalArt",
			"DigitalArt.CollectionPublicPath",
			5,
		)
		result, err := b.ExecuteScript(script, nil)
		require.NoError(t, err)
		assert.True(t, result.Reverted())
	})

}

func TestTransferDigitalArt(t *testing.T) {
	b := newBlockchain()

	accountKeys := test.AccountKeyGenerator()

	contractsObj := deployNFTContracts(t, b)

	senderAddress, senderSigner := createAccount(t, b)
	setupAccount(t, b, senderAddress, senderSigner, contractsObj)

	metadata := &iinft.Metadata{
		MetadataLink:       "QmMetadata",
		Name:               "Pure Art",
		Artist:             "Arty",
		ArtistAddress:      senderAddress,
		Description:        "Digital art in its purest form",
		Type:               "Image",
		ContentLink:        "QmContent",
		ContentPreviewLink: "QmPreview",
		Mimetype:           "image/jpeg",
		MaxEdition:         4,
		Asset:              "did:sequel:asset-id",
		Record:             "record-id",
		AssetHead:          "asset-head-id",
	}

	receiverAccountKey, receiverSigner := accountKeys.NewWithSigner()
	receiverAddress, err := b.CreateAccount([]*flow.AccountKey{receiverAccountKey}, nil)
	require.NoError(t, err)

	script, err := templates.GenerateSealDigitalArtScript(contractsObj.NFTAddress, contractsObj.DigitalArtAddress,
		metadata)
	require.NoError(t, err)

	tx := createTxWithTemplateAndAuthorizer(b, script, contractsObj.DigitalArtAddress)

	signAndSubmit(
		t, b, tx,
		[]flow.Address{
			b.ServiceKey().Address,
			contractsObj.DigitalArtAddress,
		},
		[]crypto.Signer{
			b.ServiceKey().Signer(),
			contractsObj.DigitalArtSigner,
		},
		false,
	)

	script, err = templates.GenerateMintNFTScript(metadata.Asset, contractsObj.NFTAddress, contractsObj.DigitalArtAddress,
		"DigitalArt", "DigitalArt.AdminStoragePath",
		"DigitalArt.CollectionPublicPath", senderAddress)
	require.NoError(t, err)

	tx = createTxWithTemplateAndAuthorizer(b, script, contractsObj.DigitalArtAddress)

	signAndSubmit(
		t, b, tx,
		[]flow.Address{
			b.ServiceKey().Address,
			contractsObj.DigitalArtAddress,
		},
		[]crypto.Signer{
			b.ServiceKey().Signer(),
			contractsObj.DigitalArtSigner,
		},
		false,
	)

	// create a new Collection
	t.Run("Should be able to create a new empty NFT Collection", func(t *testing.T) {

		script, err := templates.GenerateCreateCollectionScript(
			contractsObj.NFTAddress,
			contractsObj.DigitalArtAddress,
			"DigitalArt",
			"DigitalArt.CollectionStoragePath",
			"DigitalArt.CollectionPublicPath",
		)
		require.NoError(t, err)

		tx := createTxWithTemplateAndAuthorizer(b, script, receiverAddress)

		signAndSubmit(
			t, b, tx,
			[]flow.Address{
				b.ServiceKey().Address,
				receiverAddress,
			},
			[]crypto.Signer{
				b.ServiceKey().Signer(),
				receiverSigner,
			},
			false,
		)

		script = templates.GenerateInspectCollectionLenScript(
			contractsObj.NFTAddress,
			contractsObj.DigitalArtAddress,
			receiverAddress,
			"DigitalArt",
			"DigitalArt.CollectionPublicPath",
			0,
		)
		executeScriptAndCheck(t, b, script, nil)

	})

	t.Run("Shouldn't be able to withdraw an NFT that doesn't exist in a collection", func(t *testing.T) {

		script, err = templates.GenerateTransferScript(
			contractsObj.NFTAddress,
			contractsObj.DigitalArtAddress,
			"DigitalArt",
			"DigitalArt.CollectionStoragePath",
			"DigitalArt.CollectionPublicPath",
			receiverAddress,
			3,
		)
		require.NoError(t, err)
		tx := createTxWithTemplateAndAuthorizer(b, script, senderAddress)

		signAndSubmit(
			t, b, tx,
			[]flow.Address{
				b.ServiceKey().Address,
				senderAddress,
			},
			[]crypto.Signer{
				b.ServiceKey().Signer(),
				senderSigner,
			},
			true,
		)

		script = templates.GenerateInspectCollectionLenScript(
			contractsObj.NFTAddress,
			contractsObj.DigitalArtAddress,
			receiverAddress,
			"DigitalArt",
			"DigitalArt.CollectionPublicPath",
			0,
		)
		executeScriptAndCheck(t, b, script, nil)

		// Assert that the account's collection is correct
		script = templates.GenerateInspectCollectionLenScript(
			contractsObj.NFTAddress,
			contractsObj.DigitalArtAddress,
			senderAddress,
			"DigitalArt",
			"DigitalArt.CollectionPublicPath",
			1,
		)

		executeScriptAndCheck(t, b, script, nil)

	})

	// transfer an NFT
	t.Run("Should be able to withdraw an NFT and deposit to another accounts collection", func(t *testing.T) {
		script, err = templates.GenerateTransferScript(
			contractsObj.NFTAddress,
			contractsObj.DigitalArtAddress,
			"DigitalArt",
			"DigitalArt.CollectionStoragePath",
			"DigitalArt.CollectionPublicPath",
			receiverAddress,
			0,
		)
		require.NoError(t, err)

		tx := createTxWithTemplateAndAuthorizer(b, script, senderAddress)

		signAndSubmit(
			t, b, tx,
			[]flow.Address{
				b.ServiceKey().Address,
				senderAddress,
			},
			[]crypto.Signer{
				b.ServiceKey().Signer(),
				senderSigner,
			},
			false,
		)

		// Assert that the account's collection is correct
		script = templates.GenerateInspectCollectionScript(
			contractsObj.NFTAddress,
			contractsObj.DigitalArtAddress,
			receiverAddress,
			"DigitalArt",
			"DigitalArt.CollectionPublicPath",
			0,
		)
		executeScriptAndCheck(t, b, script, nil)

		script = templates.GenerateInspectCollectionLenScript(
			contractsObj.NFTAddress,
			contractsObj.DigitalArtAddress,
			receiverAddress,
			"DigitalArt",
			"DigitalArt.CollectionPublicPath",
			1,
		)
		executeScriptAndCheck(t, b, script, nil)

		// Assert that the account's collection is correct
		script = templates.GenerateInspectCollectionLenScript(
			contractsObj.NFTAddress,
			contractsObj.DigitalArtAddress,
			senderAddress,
			"DigitalArt",
			"DigitalArt.CollectionPublicPath",
			0,
		)
		executeScriptAndCheck(t, b, script, nil)

	})

	// transfer an NFT
	t.Run("Should be able to withdraw an NFT and destroy it, not reducing the supply", func(t *testing.T) {

		script := templates.GenerateDestroyScript(
			contractsObj.NFTAddress,
			contractsObj.DigitalArtAddress,
			"DigitalArt",
			"DigitalArt.CollectionStoragePath",
			0,
		)
		tx := createTxWithTemplateAndAuthorizer(b, script, receiverAddress)

		signAndSubmit(
			t, b, tx,
			[]flow.Address{
				b.ServiceKey().Address,
				receiverAddress,
			},
			[]crypto.Signer{
				b.ServiceKey().Signer(),
				receiverSigner,
			},
			false,
		)

		script = templates.GenerateInspectCollectionLenScript(
			contractsObj.NFTAddress,
			contractsObj.DigitalArtAddress,
			receiverAddress,
			"DigitalArt",
			"DigitalArt.CollectionPublicPath",
			0,
		)
		executeScriptAndCheck(t, b, script, nil)

		// Assert that the account's collection is correct
		script = templates.GenerateInspectCollectionLenScript(
			contractsObj.NFTAddress,
			contractsObj.DigitalArtAddress,
			senderAddress,
			"DigitalArt",
			"DigitalArt.CollectionPublicPath",
			0,
		)
		executeScriptAndCheck(t, b, script, nil)

		script = templates.GenerateInspectNFTSupplyScript(
			contractsObj.NFTAddress,
			contractsObj.DigitalArtAddress,
			"DigitalArt",
			1,
		)
		executeScriptAndCheck(t, b, script, nil)

	})
}
