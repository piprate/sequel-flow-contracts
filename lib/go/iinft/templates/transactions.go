package templates

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"

	"github.com/onflow/flow-go-sdk"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"

	_ "github.com/kevinburke/go-bindata"
)

//go:embed templates
var templateFS embed.FS
var goTemplates *template.Template

func init() {
	var err error
	goTemplates, err = template.New("").ParseFS(templateFS, "templates/transactions/*.cdc", "templates/scripts/*.cdc")
	if err != nil {
		panic(err)
	}
}

// GenerateCreateCollectionScript Creates a script that instantiates a new
// NFT collection instance, stores the collection in memory, then stores a
// reference to the collection.
func GenerateCreateCollectionScript(nftAddr, tokenAddr flow.Address, tokenContractName, storageLocation, publicLocation string) ([]byte, error) {
	buf := &bytes.Buffer{}
	if err := goTemplates.ExecuteTemplate(buf, "account_setup", map[string]interface{}{
		"NFTAddress":         nftAddr.String(),
		"TokenName":          tokenContractName,
		"TokenAddress":       tokenAddr.String(),
		"PrivateStoragePath": storageLocation,
		"PublicStoragePath":  publicLocation,
	}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GenerateSealDigitalArtScript Creates a script that uses the admin resource
// to mint a new NFT and deposit it into a user's collection
func GenerateSealDigitalArtScript(nftAddr, tokenAddr flow.Address, metadata *iinft.Metadata) ([]byte, error) {
	buf := &bytes.Buffer{}
	if err := goTemplates.ExecuteTemplate(buf, "master_seal", map[string]interface{}{
		"NFTAddress":       nftAddr.String(),
		"TokenName":        "DigitalArt",
		"TokenAddress":     tokenAddr.String(),
		"AdminStoragePath": "DigitalArt.AdminStoragePath",
		"MD":               metadata,
	}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GenerateMintNFTScript Creates a script that uses the admin resource
// to mint a new NFT and deposit it into a user's collection
func GenerateMintNFTScript(masterID string, nftAddr, tokenAddr flow.Address, tokenContractName, adminStorageLocation, publicLocation string, receiverAddr flow.Address) ([]byte, error) {
	buf := &bytes.Buffer{}
	if err := goTemplates.ExecuteTemplate(buf, "token_mint", map[string]interface{}{
		"NFTAddress":               nftAddr.String(),
		"TokenName":                tokenContractName,
		"TokenAddress":             tokenAddr.String(),
		"AdminStoragePath":         adminStorageLocation,
		"Master":                   masterID,
		"ReceiverAddress":          receiverAddr,
		"ReceiverPublicCollection": publicLocation,
	}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GenerateTransferScript creates a script that withdraws an NFT token
// from a collection and deposits it to another collection
func GenerateTransferScript(nftAddr, tokenAddr flow.Address, tokenContractName, storageLocation, publicLocation string, receiverAddr flow.Address, transferNFTID int) ([]byte, error) {
	buf := &bytes.Buffer{}
	if err := goTemplates.ExecuteTemplate(buf, "token_transfer", map[string]interface{}{
		"NFTAddress":                nftAddr.String(),
		"TokenName":                 tokenContractName,
		"TokenAddress":              tokenAddr.String(),
		"TokenID":                   transferNFTID,
		"RecipientAddress":          receiverAddr,
		"SenderStorageCollection":   storageLocation,
		"RecipientPublicCollection": publicLocation,
	}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GenerateDestroyScript creates a script that withdraws an NFT token
// from a collection and destroys it
func GenerateDestroyScript(nftAddr, tokenAddr flow.Address, tokenContractName, storageLocation string, destroyNFTID int) []byte {
	template := `
		import NonFungibleToken from 0x%s
		import %s from 0x%s
		transaction {
		  prepare(acct: AuthAccount) {
			let collection <- acct.load<@%s.Collection>(from:%s)!
			let nft <- collection.withdraw(withdrawID: %d)
			destroy nft
			
			acct.save(<-collection, to: %s)
		  }
		}
	`

	return []byte(fmt.Sprintf(template, nftAddr, tokenContractName, tokenAddr.String(), tokenContractName, storageLocation, destroyNFTID, storageLocation))
}
