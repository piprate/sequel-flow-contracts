{{ define "marketplace_buy_flow" }}
import NonFungibleToken from {{.NonFungibleToken}}
import FungibleToken from {{.FungibleToken}}
import NFTStorefront from {{.NFTStorefront}}
import FlowToken from {{.FlowToken}}
import DigitalArt from {{.DigitalArt}}
import SequelMarketplace from {{.SequelMarketplace}}

transaction(listingID: UInt64, storefrontAddress: Address, metadataLink: String?) {
    let listing: &{NFTStorefront.ListingPublic}
    let paymentVault: @{FungibleToken.Vault}
    let storefront: &NFTStorefront.Storefront
    let tokenReceiver: &{NonFungibleToken.Receiver}
    let buyerAddress: Address

    prepare(acct: auth(BorrowValue, SaveValue, IssueStorageCapabilityController, PublishCapability) &Account) {
        self.storefront = getAccount(storefrontAddress).capabilities.borrow<&NFTStorefront.Storefront>(NFTStorefront.StorefrontPublicPath)
            ?? panic("Could not borrow Storefront from provided address")

        self.listing = self.storefront.borrowListing(listingResourceID: listingID)
                    ?? panic("No Offer with that ID in Storefront")
        let price = self.listing.getDetails().salePrice

        let mainVault = acct.storage.borrow<auth(FungibleToken.Withdraw) &FlowToken.Vault>(from: /storage/flowTokenVault)
            ?? panic("Cannot borrow FlowToken vault from acct storage")
        self.paymentVault <- mainVault.withdraw(amount: price)

        if acct.storage.borrow<&DigitalArt.Collection>(from: DigitalArt.CollectionStoragePath) == nil {
            let collection <- DigitalArt.createEmptyCollection(nftType: Type<@DigitalArt.NFT>())
            acct.storage.save(<-collection, to: DigitalArt.CollectionStoragePath)
            let collectionCap = acct.capabilities.storage.issue<&DigitalArt.Collection>(DigitalArt.CollectionStoragePath)
            acct.capabilities.publish(collectionCap, at: DigitalArt.CollectionPublicPath)
        }

        self.tokenReceiver = acct.capabilities
            .borrow<&{NonFungibleToken.Receiver}>(DigitalArt.CollectionPublicPath)
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
            buyerAddress: self.buyerAddress,
            metadataLink: metadataLink
        )
        self.tokenReceiver.deposit(token: <-item)
    }
}
{{ end }}
