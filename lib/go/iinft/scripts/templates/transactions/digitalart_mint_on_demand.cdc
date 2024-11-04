{{ define "digitalart_mint_on_demand" }}
import NonFungibleToken from {{.NonFungibleToken}}
import FungibleToken from {{.FungibleToken}}
import FungibleTokenMetadataViews from {{.FungibleTokenMetadataViews}}
import Evergreen from {{.Evergreen}}
import DigitalArt from {{.DigitalArt}}
import SequelMarketplace from {{.SequelMarketplace}}

transaction(masterId: String, numEditions: UInt64, unitPrice: UFix64, ftContractAddress: Address, ftContractName: String, modID: UInt64) {
    let admin: &DigitalArt.Admin
    let evergreenProfile: Evergreen.Profile
    let paymentVault: @{FungibleToken.Vault}
    let tokenReceiver: &{NonFungibleToken.Receiver}
    let buyerAddress: Address
    let sellerVaultPath: PublicPath

    prepare(buyer: auth(BorrowValue, IssueStorageCapabilityController, PublishCapability, SaveValue) &Account, platform: auth(BorrowValue) &Account) {
        if numEditions == 0 {
            panic("no editions requested")
        }

        self.admin = platform.storage.borrow<&DigitalArt.Admin>(from: DigitalArt.AdminStoragePath)!

        {{- if .Parameters.Metadata }}
        if !self.admin.isSealed(masterId: masterId) {
            let metadata = DigitalArt.Metadata(
                name: {{safe .Parameters.Metadata.Name}},
                artist: {{safe .Parameters.Metadata.Artist}},
                description: {{safe .Parameters.Metadata.Description}},
                type: {{safe .Parameters.Metadata.Type}},
                contentURI: {{safe .Parameters.Metadata.ContentURI}},
                contentPreviewURI: {{safe .Parameters.Metadata.ContentPreviewURI}},
                mimetype: {{safe .Parameters.Metadata.ContentMimetype}},
                edition: {{.Parameters.Metadata.Edition}},
                maxEdition: {{.Parameters.Metadata.MaxEdition}},
                asset: masterId,
                metadataURI: {{safe .Parameters.Metadata.MetadataURI}},
                record: {{safe .Parameters.Metadata.Record}},
                assetHead: {{safe .Parameters.Metadata.AssetHead}}
            )
            let evergreenProfile = Evergreen.Profile(
                id: {{safe .Parameters.Profile.ID}},
                description: {{safe .Parameters.Profile.Description}},
                roles: [
                {{- $last := dec (len .Parameters.Profile.Roles)}}
                {{- range $i, $role := .Parameters.Profile.Roles}}
                    Evergreen.Role(id: {{safe $role.ID}},
                       description: {{safe $role.Description}},
                       initialSaleCommission: {{ufix64 $role.InitialSaleCommission}},
                       secondaryMarketCommission: {{ufix64 $role.SecondaryMarketCommission}},
                       address: 0x{{$role.Address}},
                       receiverPath: {{if $role.ReceiverPath}}{{$role.ReceiverPath}}{{else}}nil{{end}}
                    ){{ if ne $i $last}},{{ end }}
                {{end}}
                ]
            )
            self.admin.sealMaster(metadata: metadata, evergreenProfile: evergreenProfile)
        }
        {{- end}}

        if numEditions > self.admin.availableEditions(masterId: masterId) {
            panic("too many editions requested")
        }

        self.evergreenProfile = self.admin.evergreenProfile(masterId: masterId)

        // Borrow a reference to the vault stored on the passed account at the passed publicPath
        let resolverRef = getAccount(ftContractAddress)
            .contracts.borrow<&{FungibleToken}>(name: ftContractName)
                ?? panic("Could not borrow FungibleToken reference to the contract. Make sure the provided contract name ("
                          .concat(ftContractName).concat(") and address (").concat(ftContractAddress.toString()).concat(") are correct!"))

        // Use that reference to retrieve the FTView
       let vaultData = resolverRef.resolveContractView(resourceType: nil, viewType: Type<FungibleTokenMetadataViews.FTVaultData>()) as! FungibleTokenMetadataViews.FTVaultData?
            ?? panic("Could not resolve FTVaultData view. The ".concat(ftContractName).concat(" contract at ")
                .concat(ftContractAddress.toString()).concat(" needs to implement the FTVaultData Metadata view in order to execute this transaction."))

        let vaultRef = buyer.storage.borrow<auth(FungibleToken.Withdraw) &{FungibleToken.Provider}>(from: vaultData.storagePath)
            ?? panic("Cannot borrow fungible token vault from acct storage")
        let price = unitPrice * UFix64(numEditions)
        self.paymentVault <- vaultRef.withdraw(amount: price)
        self.sellerVaultPath = vaultData.receiverPath

        if buyer.storage.borrow<&DigitalArt.Collection>(from: DigitalArt.CollectionStoragePath) == nil {
            let collection <- DigitalArt.createEmptyCollection(nftType: Type<@DigitalArt.NFT>())
            buyer.storage.save(<-collection, to: DigitalArt.CollectionStoragePath)
            let collectionCap = buyer.capabilities.storage.issue<&DigitalArt.Collection>(DigitalArt.CollectionStoragePath)
            buyer.capabilities.publish(collectionCap, at: DigitalArt.CollectionPublicPath)
        }

        self.tokenReceiver = buyer.capabilities
            .borrow<&{NonFungibleToken.CollectionPublic,NonFungibleToken.Receiver}>(DigitalArt.CollectionPublicPath)
            ?? panic("Cannot borrow NFT collection receiver from acct")

        self.buyerAddress = buyer.address
    }

    execute {
        var i = UInt64(0)
        while i < numEditions {
            self.tokenReceiver.deposit(token:<- self.admin.mintEditionNFT(masterId: masterId, modID: modID))
            i = i + 1
        }

        SequelMarketplace.payForMintedTokens(
            unitPrice: unitPrice,
            numEditions: numEditions,
            sellerRole: "Artist",
            sellerVaultPath: self.sellerVaultPath,
            paymentVault: <-self.paymentVault,
            evergreenProfile: self.evergreenProfile,
        )
    }
}
{{ end }}
