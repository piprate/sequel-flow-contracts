{{ define "marketplace_buy_fusd" }}
import NonFungibleToken from {{.NonFungibleToken}}
import FungibleToken from {{.FungibleToken}}
import NFTStorefront from {{.NFTStorefront}}
import FUSD from {{.FUSD}}
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

        let mainVault = acct.borrow<&FUSD.Vault>(from: /storage/fusdVault)
            ?? panic("Cannot borrow FUSD vault from acct storage")
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
        let item <- SequelMarketplace.buyToken(
            storefrontAddress: storefrontAddress,
            storefront: self.storefront,
            listingID: listingID,
            listing: self.listing,
            paymentVault: <- self.paymentVault,
            buyerAddress: self.buyerAddress
        )
        self.tokenReceiver.deposit(token: <-item)
    }
}
{{ end }}