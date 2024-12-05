package test

import "fmt"

// inspectNFTSupplyScript creates a script that reads
// the total supply of tokens in existence
// and makes assertions about the number
func inspectNFTSupplyScript(addrMap map[string]string, tokenContractName string, expectedSupply int) string {
	template := `
		import NonFungibleToken from %s
		import %s from %s

		access(all) fun main() {
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

		access(all) fun main() {
			let acct = getAccount(0x%s)
			let collectionRef = acct.capabilities.borrow<&{NonFungibleToken.CollectionPublic}>(%s)
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

		access(all) fun main() {
			let acct = getAccount(0x%s)
			let collectionRef = acct.capabilities.borrow<&{NonFungibleToken.CollectionPublic}>(%s)
				?? panic("Could not borrow capability from public collection")
			
			let tokenRef = collectionRef.borrowNFT(UInt64(%d))
		}
	`

	return fmt.Sprintf(template, addrMap["NonFungibleToken"], tokenContractName, addrMap[tokenContractName], userAddr, publicLocation, nftID)
}
