{{ define "marketplace_buy" }}
import NonFungibleToken from {{.NonFungibleToken}}
import FungibleToken from {{.FungibleToken}}
import FungibleTokenMetadataViews from {{.FungibleTokenMetadataViews}}
import NFTStorefront from {{.NFTStorefront}}
import DigitalArt from {{.DigitalArt}}
import SequelMarketplace from {{.SequelMarketplace}}

transaction(listingID: UInt64, storefrontAddress: Address, ftContractAddress: Address, ftContractName: String, metadataLink: String?) {
    let listing: &{NFTStorefront.ListingPublic}
    let paymentVault: @{FungibleToken.Vault}
    let storefront: &NFTStorefront.Storefront
    let tokenReceiver: &{NonFungibleToken.Receiver}
    let buyerAddress: Address
    let vaultData: FungibleTokenMetadataViews.FTVaultData

    prepare(acct: auth(BorrowValue, SaveValue, IssueStorageCapabilityController, PublishCapability) &Account) {
        self.storefront = getAccount(storefrontAddress).capabilities.borrow<&NFTStorefront.Storefront>(NFTStorefront.StorefrontPublicPath)
            ?? panic("Could not borrow Storefront from provided address")

        self.listing = self.storefront.borrowListing(listingResourceID: listingID)
                    ?? panic("No Offer with that ID in Storefront")
        let price = self.listing.getDetails().salePrice

        // Borrow a reference to the vault stored on the passed account at the passed publicPath
        let resolverRef = getAccount(ftContractAddress)
            .contracts.borrow<&{FungibleToken}>(name: ftContractName)
                ?? panic("Could not borrow FungibleToken reference to the contract. Make sure the provided contract name ("
                          .concat(ftContractName).concat(") and address (").concat(ftContractAddress.toString()).concat(") are correct!"))

        // Use that reference to retrieve the FTView
        self.vaultData = resolverRef.resolveContractView(resourceType: nil, viewType: Type<FungibleTokenMetadataViews.FTVaultData>()) as! FungibleTokenMetadataViews.FTVaultData?
            ?? panic("Could not resolve FTVaultData view. The ".concat(ftContractName).concat(" contract at ")
                .concat(ftContractAddress.toString()).concat(" needs to implement the FTVaultData Metadata view in order to execute this transaction."))

        let vaultRef = acct.storage.borrow<auth(FungibleToken.Withdraw) &{FungibleToken.Provider}>(from: self.vaultData.storagePath)
            ?? panic("Cannot borrow fungible token vault from acct storage")
        self.paymentVault <- vaultRef.withdraw(amount: price)

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
