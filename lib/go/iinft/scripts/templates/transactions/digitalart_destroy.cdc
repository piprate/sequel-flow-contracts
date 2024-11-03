{{ define "digitalart_destroy" }}
import NonFungibleToken from {{.NonFungibleToken}}
import Burner from {{.Burner}}
import DigitalArt from {{.DigitalArt}}

transaction(tokenId: UInt64) {
    let collectionRef: auth(NonFungibleToken.Withdraw) &DigitalArt.Collection

    prepare(signer: auth(BorrowValue) &Account) {
        self.collectionRef = signer.storage.borrow<auth(NonFungibleToken.Withdraw) &DigitalArt.Collection>(from: DigitalArt.CollectionStoragePath)!
    }

    execute {
        // withdraw the NFT from the owner's collection
        let nft <- self.collectionRef.withdraw(withdrawID: tokenId)

        Burner.burn(<-nft)
    }
}
{{ end }}
