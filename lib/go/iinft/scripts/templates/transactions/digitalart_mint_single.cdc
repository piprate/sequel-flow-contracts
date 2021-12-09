{{ define "digitalart_mint_single" }}
import NonFungibleToken from {{.NonFungibleToken}}
import DigitalArt from {{.DigitalArt}}

transaction(metadataLink: String,
            name: String,
            artist: String,
            description: String,
            type: String,
            contentLink: String,
            contentPreviewLink: String,
            mimetype: String,
            asset: String,
            record: String,
            assetHead: String,
            participationProfileID: UInt32,
            artistAddress: Address?,
            artistInitial: UFix64,
            artistSecondary: UFix64,
            platformAddress: Address?,
            platformInitial: UFix64,
            platformSecondary: UFix64,
            recipientAddr: Address) {

    let admin: &DigitalArt.Admin

    prepare(signer: AuthAccount) {
        self.admin = signer.borrow<&DigitalArt.Admin>(from: DigitalArt.AdminStoragePath)!
    }

    execute {
        let recipient = getAccount(recipientAddr)

        let receiver = recipient
            .getCapability(DigitalArt.CollectionPublicPath)!
            .borrow<&{DigitalArt.CollectionPublic}>()
            ?? panic("Could not get receiver reference to the NFT Collection")

        let roles: { String: DigitalArt.ParticipationRole } = {}
        if artistAddress != nil {
            roles["Artist"] = DigitalArt.ParticipationRole(
                id: "Artist",
                initialSaleCommission: artistInitial,
                secondaryMarketCommission: artistSecondary,
                address: artistAddress!
            )
        }
        if platformAddress != nil {
            roles["Platform"] = DigitalArt.ParticipationRole(
                id: "Platform",
                initialSaleCommission: platformInitial,
                secondaryMarketCommission: platformSecondary,
                address: platformAddress!
            )
        }

        let newNFT <- self.admin.mintSingleNFT(metadata: DigitalArt.Metadata(
            metadataLink: metadataLink,
            name: name,
            artist: artist,
            description: description,
            type: type,
            contentLink: contentLink,
            contentPreviewLink: contentPreviewLink,
            mimetype: mimetype,
            edition: 1,
            maxEdition: 1,
            asset: asset,
            record: record,
            assetHead: assetHead,
            participationProfile: DigitalArt.ParticipationProfile(
                id: participationProfileID,
                roles: roles,
                description: ""
            )
        ))
        receiver.deposit(token: <-newNFT)
    }
}
{{ end }}