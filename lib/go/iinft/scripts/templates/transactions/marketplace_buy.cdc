{{ define "marketplace_buy" }}
import NonFungibleToken from {{.NonFungibleToken}}
import FungibleToken from {{.FungibleToken}}
import NFTStorefront from {{.NFTStorefront}}
import FlowToken from {{.FlowToken}}
import DigitalArt from {{.DigitalArt}}
import SequelMarketplace from {{.SequelMarketplace}}

transaction(listingID: UInt64, storefrontAddress: Address) {
    let listing: &NFTStorefront.Listing{NFTStorefront.ListingPublic}
    let paymentVault: @FungibleToken.Vault
    let storefront: &NFTStorefront.Storefront{NFTStorefront.StorefrontPublic}
    let tokenReceiver: &{NonFungibleToken.CollectionPublic,NonFungibleToken.Receiver}
    let buyerAddress: Address

    prepare(acct: AuthAccount) {
        self.storefront = getAccount(storefrontAddress)
            .getCapability(NFTStorefront.StorefrontPublicPath)!
            .borrow<&NFTStorefront.Storefront{NFTStorefront.StorefrontPublic}>()
            ?? panic("Could not borrow Storefront from provided address")

        self.listing = self.storefront.borrowListing(listingResourceID: listingID)
                    ?? panic("No Offer with that ID in Storefront")
        let price = self.listing.getDetails().salePrice

        let mainVault = acct.borrow<&FlowToken.Vault>(from: /storage/flowTokenVault)
            ?? panic("Cannot borrow FlowToken vault from acct storage")
        self.paymentVault <- mainVault.withdraw(amount: price)

        if acct.borrow<&DigitalArt.Collection>(from: DigitalArt.CollectionStoragePath) == nil {
            let collection <- DigitalArt.createEmptyCollection() as! @DigitalArt.Collection
            acct.save(<-collection, to: DigitalArt.CollectionStoragePath)
            acct.link<&{NonFungibleToken.CollectionPublic,NonFungibleToken.Receiver}>(DigitalArt.CollectionPublicPath, target: DigitalArt.CollectionStoragePath)
        }

        self.tokenReceiver = acct.getCapability(DigitalArt.CollectionPublicPath)
            .borrow<&{NonFungibleToken.CollectionPublic,NonFungibleToken.Receiver}>()
            ?? panic("Cannot borrow NFT collection receiver from acct")

        self.buyerAddress = acct.address
    }

    execute {
        let item <- self.listing.purchase(payment: <-self.paymentVault)
        self.storefront.cleanup(listingResourceID: listingID)
        self.tokenReceiver.deposit(token: <-item)
    }
}
{{ end }}