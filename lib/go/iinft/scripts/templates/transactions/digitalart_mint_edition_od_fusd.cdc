{{ define "digitalart_mint_edition_od_fusd" }}
import NonFungibleToken from {{.NonFungibleToken}}
import FungibleToken from {{.FungibleToken}}
import FUSD from {{.FUSD}}
import Evergreen from {{.Evergreen}}
import DigitalArt from {{.DigitalArt}}
import SequelMarketplace from {{.SequelMarketplace}}

transaction(masterId: String, numEditions: UInt64, unitPrice: UFix64) {
    let admin: &DigitalArt.Admin
    let availableEditions: UInt64
    let evergreenProfile: Evergreen.Profile
    let paymentVault: @FungibleToken.Vault
    let tokenReceiver: &{NonFungibleToken.CollectionPublic,NonFungibleToken.Receiver}
    let buyerAddress: Address

    prepare(buyer: AuthAccount, platform: AuthAccount) {
        self.admin = platform.borrow<&DigitalArt.Admin>(from: DigitalArt.AdminStoragePath)!
        self.availableEditions = self.admin.availableEditions(masterId: masterId)
        self.evergreenProfile = self.admin.evergreenProfile(masterId: masterId)

        let mainVault = buyer.borrow<&FUSD.Vault>(from: /storage/fusdVault)
            ?? panic("Cannot borrow FUSD vault from acct storage")
        let price = unitPrice * UFix64(numEditions)
        self.paymentVault <- mainVault.withdraw(amount: price)

        if buyer.borrow<&DigitalArt.Collection>(from: DigitalArt.CollectionStoragePath) == nil {
            let collection <- DigitalArt.createEmptyCollection() as! @DigitalArt.Collection
            buyer.save(<-collection, to: DigitalArt.CollectionStoragePath)
            buyer.link<&{NonFungibleToken.CollectionPublic,NonFungibleToken.Receiver}>(DigitalArt.CollectionPublicPath, target: DigitalArt.CollectionStoragePath)
        }

        self.tokenReceiver = buyer.getCapability(DigitalArt.CollectionPublicPath)
            .borrow<&{NonFungibleToken.CollectionPublic,NonFungibleToken.Receiver}>()
            ?? panic("Cannot borrow NFT collection receiver from acct")

        self.buyerAddress = buyer.address
    }

    execute {
        if numEditions > self.availableEditions {
            panic("too many editions requested")
        }

        if numEditions == 0 {
            return
        }

        var i = UInt64(0)
        while i < numEditions {
            self.tokenReceiver.deposit(token:<- self.admin.mintEditionNFT(masterId: masterId))
            i = i + 1
        }

        SequelMarketplace.payForMintedTokens(
            unitPrice: unitPrice,
            numEditions: numEditions,
            paymentVaultPath: /public/fusdReceiver,
            paymentVault: <-self.paymentVault,
            evergreenProfile: self.evergreenProfile,
        )
    }
}
{{ end }}