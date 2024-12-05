{{ define "digitalart_mint_edition" }}
import NonFungibleToken from {{.NonFungibleToken}}
import DigitalArt from {{.DigitalArt}}

transaction(masterId: String, amount: UInt64, recipient: Address) {
    let admin: &DigitalArt.Admin
    let availableEditions: UInt64
    /// Reference to the receiver's collection
    let recipientCollectionRef: &{NonFungibleToken.Receiver}

    prepare(signer: auth(BorrowValue) &Account) {
        self.admin = signer.storage.borrow<&DigitalArt.Admin>(from: DigitalArt.AdminStoragePath)!
        self.availableEditions = self.admin.availableEditions(masterId: masterId)

        // Borrow the recipient's public NFT collection reference
        self.recipientCollectionRef = getAccount(recipient).capabilities.borrow<&{NonFungibleToken.Receiver}>(DigitalArt.CollectionPublicPath)
            ?? panic("The recipient does not have a NonFungibleToken Receiver at "
                    .concat(DigitalArt.CollectionPublicPath.toString())
                    .concat(" that is capable of receiving an NFT.")
                    .concat("The recipient must initialize their account with this collection and receiver first!"))
    }

    execute {
        if amount > self.availableEditions {
            panic("too many editions requested")
        }

        var i = UInt64(0)
        while i < amount {
            let newNFT <- self.admin.mintEditionNFT(masterId: masterId, modID: 0)
            self.recipientCollectionRef.deposit(token: <-newNFT)
            i = i + 1
        }
    }
}
{{ end }}
