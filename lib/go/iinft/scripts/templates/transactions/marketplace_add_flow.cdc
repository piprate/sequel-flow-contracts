{{ define "marketplace_add_flow" }}
import NonFungibleToken from {{.NonFungibleToken}}
import NFTStorefront from {{.NFTStorefront}}
import FlowToken from {{.FlowToken}}
import DigitalArt from {{.DigitalArt}}
import SequelMarketplace from {{.SequelMarketplace}}

transaction(tokenID: UInt64, price: UFix64, initialSale: Bool) {
  let nftProviderCapability: Capability<&{NonFungibleToken.Provider,NonFungibleToken.CollectionPublic,DigitalArt.CollectionPublic}>
  let storefront: &NFTStorefront.Storefront

  prepare(acct: AuthAccount) {
    let nftProviderPath = /private/SequelNFTProviderForNFTStorefront
    if !acct.getCapability<&{NonFungibleToken.Provider,NonFungibleToken.CollectionPublic,DigitalArt.CollectionPublic}>(nftProviderPath)!.check() {
        acct.link<&{NonFungibleToken.Provider,NonFungibleToken.CollectionPublic,DigitalArt.CollectionPublic}>(nftProviderPath, target: DigitalArt.CollectionStoragePath)
    }

    self.nftProviderCapability = acct.getCapability<&{NonFungibleToken.Provider,NonFungibleToken.CollectionPublic,DigitalArt.CollectionPublic}>(nftProviderPath)!
    assert(self.nftProviderCapability.borrow() != nil, message: "Missing or mis-typed nft collection provider")

    if acct.borrow<&NFTStorefront.Storefront>(from: NFTStorefront.StorefrontStoragePath) == nil {
        let storefront <- NFTStorefront.createStorefront() as! @NFTStorefront.Storefront
        acct.save(<-storefront, to: NFTStorefront.StorefrontStoragePath)
        acct.link<&NFTStorefront.Storefront{NFTStorefront.StorefrontPublic}>(NFTStorefront.StorefrontPublicPath, target: NFTStorefront.StorefrontStoragePath)
    }
    self.storefront = acct.borrow<&NFTStorefront.Storefront>(from: NFTStorefront.StorefrontStoragePath)
        ?? panic("Missing or mis-typed NFTStorefront Storefront")
  }

  execute {
    SequelMarketplace.addListing(
        storefront: self.storefront,
        nftProviderCapability: self.nftProviderCapability,
        nftType: Type<@DigitalArt.NFT>(),
        nftID: tokenID,
        salePaymentVaultPath: /public/flowTokenReceiver,
        salePaymentVaultType: Type<@FlowToken.Vault>(),
        price: price,
        initialSale: initialSale,
        extraRoles: []
    )
  }
}
{{ end }}