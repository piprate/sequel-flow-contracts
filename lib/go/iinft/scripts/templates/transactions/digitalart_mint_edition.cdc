{{ define "digitalart_mint_edition" }}
import NonFungibleToken from {{.NonFungibleToken}}
import DigitalArt from {{.DigitalArt}}

transaction(masterId: String, amount: UInt64, recipientAddr: Address) {
    let admin: &DigitalArt.Admin
    let availableEditions: UInt64

    prepare(signer: AuthAccount) {
        self.admin = signer.borrow<&DigitalArt.Admin>(from: DigitalArt.AdminStoragePath)!
        self.availableEditions = self.admin.availableEditions(masterId: masterId)
    }

    execute {
        let recipient = getAccount(recipientAddr)

        let receiver = recipient
            .getCapability(DigitalArt.CollectionPublicPath)!
            .borrow<&{DigitalArt.CollectionPublic}>()
            ?? panic("Could not get receiver reference to the NFT Collection")

        if amount > self.availableEditions {
            panic("too many editions requested")
        }

        var i = UInt64(0)
        while i < amount {
            let newNFT <- self.admin.mintEditionNFT(masterId: masterId)
            receiver.deposit(token: <-newNFT)
            i = i + 1
        }
    }
}
{{ end }}