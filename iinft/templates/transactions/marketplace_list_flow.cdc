{{ define "marketplace_list_flow" }}
import NonFungibleToken from {{.NonFungibleToken}}
import NFTStorefront from {{.NFTStorefront}}
import FlowToken from {{.FlowToken}}
import Evergreen from {{.Evergreen}}
import DigitalArt from {{.DigitalArt}}
import SequelMarketplace from {{.SequelMarketplace}}

transaction(tokenID: UInt64, price: UFix64, metadataLink: String?) {
  let nftProviderCapability: Capability<auth(NonFungibleToken.Withdraw) &DigitalArt.Collection>
  let storefront: auth(NFTStorefront.CreateListing) &NFTStorefront.Storefront

  prepare(acct: auth(BorrowValue, SaveValue, IssueStorageCapabilityController, PublishCapability) &Account) {
    self.nftProviderCapability = acct.capabilities.storage.issue<auth(NonFungibleToken.Withdraw) &DigitalArt.Collection>(
        DigitalArt.CollectionStoragePath
    )
    assert(self.nftProviderCapability.check(), message: "Missing or mis-typed nft collection provider")

    // If the account doesn't already have a Storefront
    if acct.storage.borrow<&NFTStorefront.Storefront>(from: NFTStorefront.StorefrontStoragePath) == nil {

        // Create a new empty .Storefront
        let storefront <- NFTStorefront.createStorefront()

        // save it to the account
        acct.storage.save(<-storefront, to: NFTStorefront.StorefrontStoragePath)

        // create a public capability for the .Storefront & publish
        let storefrontPublicCap = acct.capabilities.storage.issue<&{NFTStorefront.StorefrontPublic}>(
            NFTStorefront.StorefrontStoragePath
        )
        acct.capabilities.publish(storefrontPublicCap, at: NFTStorefront.StorefrontPublicPath)
    }

    self.storefront = acct.storage.borrow<auth(NFTStorefront.CreateListing) &NFTStorefront.Storefront>(from: NFTStorefront.StorefrontStoragePath)
        ?? panic("Could not borrow Storefront from provided address")
  }

  execute {
    SequelMarketplace.listToken(
        storefront: self.storefront,
        nftProviderCapability: self.nftProviderCapability,
        nftType: Type<@DigitalArt.NFT>(),
        nftID: tokenID,
        sellerVaultPath: /public/flowTokenReceiver,
        paymentVaultType: Type<@FlowToken.Vault>(),
        price: price,
        extraRoles: [],
        metadataLink: metadataLink
    )
  }
}
{{ end }}
