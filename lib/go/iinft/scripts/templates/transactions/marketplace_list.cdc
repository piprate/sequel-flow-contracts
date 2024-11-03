{{ define "marketplace_list" }}
import FungibleToken from {{.FungibleToken}}
import FungibleTokenMetadataViews from {{.FungibleTokenMetadataViews}}
import NonFungibleToken from {{.NonFungibleToken}}
import Burner from {{.Burner}}
import NFTStorefront from {{.NFTStorefront}}
import Evergreen from {{.Evergreen}}
import DigitalArt from {{.DigitalArt}}
import SequelMarketplace from {{.SequelMarketplace}}

transaction(tokenID: UInt64, price: UFix64, ftContractAddress: Address, ftContractName: String, metadataLink: String?) {
  let nftProviderCapability: Capability<auth(NonFungibleToken.Withdraw) &DigitalArt.Collection>
  let storefront: auth(NFTStorefront.CreateListing) &NFTStorefront.Storefront
  // FTVaultData struct to get paths from
  let vaultData: FungibleTokenMetadataViews.FTVaultData
  let paymentVaultType: Type

  prepare(acct: auth(BorrowValue, SaveValue, IssueStorageCapabilityController, PublishCapability) &Account) {
    // Borrow a reference to the vault stored on the passed account at the passed publicPath
    let resolverRef = getAccount(ftContractAddress)
        .contracts.borrow<&{FungibleToken}>(name: ftContractName)
            ?? panic("Could not borrow FungibleToken reference to the contract. Make sure the provided contract name ("
                      .concat(ftContractName).concat(") and address (").concat(ftContractAddress.toString()).concat(") are correct!"))

    // Use that reference to retrieve the FTView
    self.vaultData = resolverRef.resolveContractView(resourceType: nil, viewType: Type<FungibleTokenMetadataViews.FTVaultData>()) as! FungibleTokenMetadataViews.FTVaultData?
        ?? panic("Could not resolve FTVaultData view. The ".concat(ftContractName).concat(" contract at ")
            .concat(ftContractAddress.toString()).concat(" needs to implement the FTVaultData Metadata view in order to execute this transaction."))

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

    // Create a new empty vault to extract vault type, instead of using
    // vaultData.receiverLinkedType which is a reference.
    let emptyVault <-self.vaultData.createEmptyVault()
    self.paymentVaultType = emptyVault.getType()
    Burner.burn(<-emptyVault)
  }

  execute {
    SequelMarketplace.listToken(
        storefront: self.storefront,
        nftProviderCapability: self.nftProviderCapability,
        nftType: Type<@DigitalArt.NFT>(),
        nftID: tokenID,
        sellerVaultPath: self.vaultData.receiverPath,
        paymentVaultType: self.paymentVaultType,
        price: price,
        extraRoles: [],
        metadataLink: metadataLink
    )
  }
}
{{ end }}
