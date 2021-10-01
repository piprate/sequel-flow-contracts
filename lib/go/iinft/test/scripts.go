package test

import (
	"fmt"
)

// inspectNFTSupplyScript creates a script that reads
// the total supply of tokens in existence
// and makes assertions about the number
func inspectNFTSupplyScript(nftAddr, tokenAddr, tokenContractName string, expectedSupply int) string {
	template := `
		import NonFungibleToken from 0x%s
		import %s from 0x%s

		pub fun main() {
			assert(
                %s.totalSupply == UInt64(%d),
                message: "incorrect totalSupply!"
            )
		}
	`

	return fmt.Sprintf(template, nftAddr, tokenContractName, tokenAddr, tokenContractName, expectedSupply)
}

// inspectCollectionLenScript creates a script that retrieves an NFT collection
// from storage and tries to borrow a reference for an NFT that it owns.
// If it owns it, it will not fail.
func inspectCollectionLenScript(nftAddr, tokenAddr, userAddr, tokenContractName, publicLocation string, length int) string {
	template := `
		import NonFungibleToken from 0x%s
		import %s from 0x%s

		pub fun main() {
			let acct = getAccount(0x%s)
			let collectionRef = acct.getCapability(%s)!.borrow<&{NonFungibleToken.CollectionPublic}>()
				?? panic("Could not borrow capability from public collection")
			
			if %d != collectionRef.getIDs().length {
				panic("Collection Length is not correct")
			}
		}
	`

	return fmt.Sprintf(template, nftAddr, tokenContractName, tokenAddr, userAddr, publicLocation, length)
}

// inspectCollectionScript creates a script that retrieves an NFT collection
// from storage and tries to borrow a reference for an NFT that it owns.
// If it owns it, it will not fail.
func inspectCollectionScript(nftAddr, tokenAddr, userAddr, tokenContractName, publicLocation string, nftID int) string {
	template := `
		import NonFungibleToken from 0x%s
		import %s from 0x%s

		pub fun main() {
			let acct = getAccount(0x%s)
			let collectionRef = acct.getCapability(%s)!.borrow<&{NonFungibleToken.CollectionPublic}>()
				?? panic("Could not borrow capability from public collection")
			
			let tokenRef = collectionRef.borrowNFT(id: UInt64(%d))
		}
	`

	return fmt.Sprintf(template, nftAddr, tokenContractName, tokenAddr, userAddr, publicLocation, nftID)
}
