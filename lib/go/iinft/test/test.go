package test

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-emulator"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	sdktemplates "github.com/onflow/flow-go-sdk/templates"
	"github.com/onflow/flow-go-sdk/test"
	nftcontracts "github.com/onflow/flow-nft/lib/go/contracts"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/contracts"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/onflow/flow-go-sdk"
)

const (
	flowTokenName         = "FlowToken"
	nonFungibleTokenName  = "NonFungibleToken"
	defaultAccountFunding = "1000.0"
)

type Contracts struct {
	NFTAddress        flow.Address
	DigitalArtAddress flow.Address
	DigitalArtSigner  crypto.Signer
}

func deployNFTContracts(t *testing.T, b *emulator.Blockchain) Contracts {
	nftCode := nftcontracts.NonFungibleToken()
	nftAddress, err := b.CreateAccount(nil,
		[]sdktemplates.Contract{
			{
				Name:   nonFungibleTokenName,
				Source: string(nftCode),
			},
		},
	)
	require.NoError(t, err)

	_, err = b.CommitBlock()
	require.NoError(t, err)

	accountKeys := test.AccountKeyGenerator()

	digitalArtAccountKey, digitalArtSigner := accountKeys.NewWithSigner()

	digitalArtCode := contracts.GenerateDigitalArtContract(nftAddress.String())
	digitalArtAddress, err := b.CreateAccount(
		[]*flow.AccountKey{digitalArtAccountKey},
		[]sdktemplates.Contract{
			{
				Name:   "DigitalArt",
				Source: string(digitalArtCode),
			},
		},
	)
	require.NoError(t, err)

	_, err = b.CommitBlock()
	require.NoError(t, err)

	return Contracts{
		nftAddress,
		digitalArtAddress,
		digitalArtSigner,
	}
}

// newBlockchain returns an emulator blockchain for testing.
func newBlockchain(opts ...emulator.Option) *emulator.Blockchain {
	b, err := emulator.NewBlockchain(
		append(
			[]emulator.Option{
				emulator.WithStorageLimitEnabled(false),
			},
			opts...,
		)...,
	)
	if err != nil {
		panic(err)
	}
	return b
}

func createTxWithTemplateAndAuthorizer(
	b *emulator.Blockchain,
	script []byte,
	authorizerAddress flow.Address,
) *flow.Transaction {

	tx := flow.NewTransaction().
		SetScript(script).
		SetGasLimit(9999).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(authorizerAddress)

	return tx
}

// signAndSubmit signs a transaction with an array of signers and adds their signatures to the transaction
// before submitting it to the emulator.
//
// If the private keys do not match up with the addresses, the transaction will not succeed.
//
// The shouldRevert parameter indicates whether the transaction should fail or not.
//
// This function asserts the correct result and commits the block if it passed.
func signAndSubmit(
	t *testing.T,
	b *emulator.Blockchain,
	tx *flow.Transaction,
	signerAddresses []flow.Address,
	signers []crypto.Signer,
	shouldRevert bool,
) {
	// sign transaction with each signer
	for i := len(signerAddresses) - 1; i >= 0; i-- {
		signerAddress := signerAddresses[i]
		signer := signers[i]

		if i == 0 {
			err := tx.SignEnvelope(signerAddress, 0, signer)
			assert.NoError(t, err)
		} else {
			err := tx.SignPayload(signerAddress, 0, signer)
			assert.NoError(t, err)
		}
	}

	Submit(t, b, tx, shouldRevert)
}

// Submit submits a transaction and checks if it fails or not.
func Submit(
	t *testing.T,
	b *emulator.Blockchain,
	tx *flow.Transaction,
	shouldRevert bool,
) {
	// submit the signed transaction
	err := b.AddTransaction(*tx)
	require.NoError(t, err)

	result, err := b.ExecuteNextTransaction()
	require.NoError(t, err)

	if shouldRevert {
		assert.True(t, result.Reverted())
	} else {
		if !assert.True(t, result.Succeeded()) {
			t.Log(result.Error.Error())
		}
	}

	_, err = b.CommitBlock()
	assert.NoError(t, err)
}

// executeScriptAndCheck executes a script and checks to make sure that it succeeded.
func executeScriptAndCheck(t *testing.T, b *emulator.Blockchain, script []byte, arguments [][]byte) cadence.Value {
	result, err := b.ExecuteScript(script, arguments)
	require.NoError(t, err)
	if !assert.True(t, result.Succeeded()) {
		t.Log(result.Error.Error())
	}

	return result.Value
}

func readFile(path string) []byte {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return contents
}

// CadenceUFix64 returns a UFix64 value
func CadenceUFix64(value string) cadence.Value {
	newValue, err := cadence.NewUFix64(value)

	if err != nil {
		panic(err)
	}

	return newValue
}

func bytesToCadenceArray(b []byte) cadence.Array {
	values := make([]cadence.Value, len(b))

	for i, v := range b {
		values[i] = cadence.NewUInt8(v)
	}

	return cadence.NewArray(values)
}

// Simple error-handling wrapper for Flow account creation.
func createAccount(t *testing.T, b *emulator.Blockchain) (sdk.Address, crypto.Signer) {
	accountKeys := test.AccountKeyGenerator()
	accountKey, signer := accountKeys.NewWithSigner()

	address, err := b.CreateAccount([]*sdk.AccountKey{accountKey}, nil)
	require.NoError(t, err)

	return address, signer
}

func setupAccount(
	t *testing.T,
	b *emulator.Blockchain,
	address flow.Address,
	signer crypto.Signer,
	contracts Contracts,
) (sdk.Address, crypto.Signer) {
	setupDigitalArtCollection(t, b, address, signer, contracts.NFTAddress, contracts.DigitalArtAddress)
	//fundAccount(t, b, address, defaultAccountFunding)

	return address, signer
}

func setupDigitalArtCollection(
	t *testing.T,
	b *emulator.Blockchain,
	userAddress flow.Address,
	userSigner crypto.Signer,
	nftAddress, exampleNFTAddress flow.Address,
) {
	script, err := templates.GenerateCreateCollectionScript(
		nftAddress,
		exampleNFTAddress,
		"DigitalArt",
		"DigitalArt.CollectionStoragePath",
		"DigitalArt.CollectionPublicPath",
	)
	require.NoError(t, err)

	tx := flow.NewTransaction().
		SetScript(script).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(userAddress)

	signAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, userAddress},
		[]crypto.Signer{b.ServiceKey().Signer(), userSigner},
		false,
	)
}

// assertEqual asserts that two objects are equal.
//
//    assertEqual(t, 123, 123)
//
// Pointer variable equality is determined based on the equality of the
// referenced values (as opposed to the memory addresses). Function equality
// cannot be determined and will always fail.
//
func assertEqual(t *testing.T, expected, actual interface{}) bool {

	if assert.ObjectsAreEqual(expected, actual) {
		return true
	}

	message := fmt.Sprintf(
		"Not equal: \nexpected: %s\nactual  : %s",
		expected,
		actual,
	)

	return assert.Fail(t, message)
}
