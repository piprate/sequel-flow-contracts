{{ define "digitalart_mint_on_demand_flow" }}
import NonFungibleToken from {{.NonFungibleToken}}
import FungibleToken from {{.FungibleToken}}
import FlowToken from {{.FlowToken}}
import Evergreen from {{.Evergreen}}
import DigitalArt from {{.DigitalArt}}
import SequelMarketplace from {{.SequelMarketplace}}

transaction(masterId: String, numEditions: UInt64, unitPrice: UFix64, modID: UInt64) {
    let admin: &DigitalArt.Admin
    let evergreenProfile: Evergreen.Profile
    let paymentVault: @FungibleToken.Vault
    let tokenReceiver: &{NonFungibleToken.CollectionPublic,NonFungibleToken.Receiver}
    let buyerAddress: Address

    prepare(buyer: AuthAccount, platform: AuthAccount) {
        if numEditions == 0 {
            panic("no editions requested")
        }

        self.admin = platform.borrow<&DigitalArt.Admin>(from: DigitalArt.AdminStoragePath)!

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

        let mainVault = buyer.borrow<&FlowToken.Vault>(from: /storage/flowTokenVault)
            ?? panic("Cannot borrow FlowToken vault from acct storage")
        let price = unitPrice * UFix64(numEditions)
        self.paymentVault <- mainVault.withdraw(amount: price)

        if buyer.borrow<&DigitalArt.Collection>(from: DigitalArt.CollectionStoragePath) == nil {
            let collection <- DigitalArt.createEmptyCollection() as! @DigitalArt.Collection
            buyer.save(<-collection, to: DigitalArt.CollectionStoragePath)
            buyer.link<&{NonFungibleToken.CollectionPublic,NonFungibleToken.Receiver, DigitalArt.CollectionPublic}>(
                DigitalArt.CollectionPublicPath,
                target: DigitalArt.CollectionStoragePath
            )
        }

        self.tokenReceiver = buyer.getCapability(DigitalArt.CollectionPublicPath)
            .borrow<&{NonFungibleToken.CollectionPublic,NonFungibleToken.Receiver}>()
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
            sellerVaultPath: /public/flowTokenReceiver,
            paymentVault: <-self.paymentVault,
            evergreenProfile: self.evergreenProfile,
        )
    }
}
{{ end }}