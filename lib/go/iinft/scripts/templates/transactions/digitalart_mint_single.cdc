{{ define "digitalart_mint_single" }}
import NonFungibleToken from {{.NonFungibleToken}}
import Evergreen from {{.Evergreen}}
import DigitalArt from {{.DigitalArt}}

transaction(metadata: DigitalArt.Metadata, evergreenProfile: Evergreen.Profile, recipientAddr: Address) {

    let admin: &DigitalArt.Admin

    prepare(signer: AuthAccount) {
        self.admin = signer.borrow<&DigitalArt.Admin>(from: DigitalArt.AdminStoragePath)!
    }

    execute {
        let recipient = getAccount(recipientAddr)

        let receiver = recipient
            .getCapability(DigitalArt.CollectionPublicPath)!
            .borrow<&{DigitalArt.CollectionPublic}>()
            ?? panic("Could not get receiver reference to the NFT Collection")

        let newNFT <- self.admin.mintSingleNFT(metadata: metadata, evergreenProfile: evergreenProfile)
        receiver.deposit(token: <-newNFT)
    }
}
{{ end }}