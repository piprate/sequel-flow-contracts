package test

import (
	"fmt"
	"testing"

	"github.com/onflow/cadence"
	fttemplates "github.com/onflow/flow-ft/lib/go/templates"
	"github.com/onflow/flow-go-sdk"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/scripts"
)

// inspectNFTSupplyScript creates a script that reads
// the total supply of tokens in existence
// and makes assertions about the number
func inspectNFTSupplyScript(addrMap map[string]string, tokenContractName string, expectedSupply int) string {
	template := `
		import NonFungibleToken from %s
		import %s from %s

		pub fun main() {
			assert(
                %s.totalSupply == UInt64(%d),
                message: "incorrect totalSupply!"
            )
		}
	`

	return fmt.Sprintf(template, addrMap["NonFungibleToken"], tokenContractName, addrMap[tokenContractName], tokenContractName, expectedSupply)
}

// inspectCollectionLenScript creates a script that retrieves an NFT collection
// from storage and tries to borrow a reference for an NFT that it owns.
// If it owns it, it will not fail.
func inspectCollectionLenScript(addrMap map[string]string, userAddr, tokenContractName, publicLocation string, length int) string {
	template := `
		import NonFungibleToken from %s
		import %s from %s

		pub fun main() {
			let acct = getAccount(0x%s)
			let collectionRef = acct.getCapability(%s)!.borrow<&{NonFungibleToken.CollectionPublic}>()
				?? panic("Could not borrow capability from public collection")
			
			if %d != collectionRef.getIDs().length {
				panic("Collection Length is not correct")
			}
		}
	`

	return fmt.Sprintf(template, addrMap["NonFungibleToken"], tokenContractName, addrMap[tokenContractName], userAddr, publicLocation, length)
}

// inspectCollectionScript creates a script that retrieves an NFT collection
// from storage and tries to borrow a reference for an NFT that it owns.
// If it owns it, it will not fail.
func inspectCollectionScript(addrMap map[string]string, userAddr, tokenContractName, publicLocation string, nftID uint64) string {
	template := `
		import NonFungibleToken from %s
		import %s from %s

		pub fun main() {
			let acct = getAccount(0x%s)
			let collectionRef = acct.getCapability(%s)!.borrow<&{NonFungibleToken.CollectionPublic}>()
				?? panic("Could not borrow capability from public collection")
			
			let tokenRef = collectionRef.borrowNFT(id: UInt64(%d))
		}
	`

	return fmt.Sprintf(template, addrMap["NonFungibleToken"], tokenContractName, addrMap[tokenContractName], userAddr, publicLocation, nftID)
}

func fundAccount(t *testing.T, se *scripts.Engine, receiverAddress flow.Address, amount string) {
	script := string(fttemplates.GenerateMintTokensScript(
		flow.HexToAddress(se.WellKnownAddresses()["FungibleToken"]),
		flow.HexToAddress(se.WellKnownAddresses()["FlowToken"]),
		"FlowToken",
	))

	_ = se.NewInlineTransaction(script).
		Argument(cadence.NewAddress(receiverAddress)).
		UFix64Argument(amount).
		SignProposeAndPayAsService().
		Test(t).
		AssertSuccess()
}
