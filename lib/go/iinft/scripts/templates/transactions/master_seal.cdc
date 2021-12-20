{{ define "master_seal" }}
import NonFungibleToken from {{.NonFungibleToken}}
import Evergreen from {{.Evergreen}}
import DigitalArt from {{.DigitalArt}}

transaction(metadataLink: String,
            name: String,
            artist: String,
            description: String,
            type: String,
            contentLink: String,
            contentPreviewLink: String,
            mimetype: String,
            maxEdition: UInt64,
            asset: String,
            record: String,
            assetHead: String,
            evergreenProfileID: UInt32,
            artistAddress: Address?,
            artistInitial: UFix64,
            artistSecondary: UFix64,
            platformAddress: Address?,
            platformInitial: UFix64,
            platformSecondary: UFix64) {
    let admin: &DigitalArt.Admin

    prepare(signer: AuthAccount) {

        self.admin = signer.borrow<&DigitalArt.Admin>(from: DigitalArt.AdminStoragePath)!
    }

    execute {
        let roles: { String: Evergreen.Role } = {}
        if artistAddress != nil {
            roles["Artist"] = Evergreen.Role(
                id: "Artist",
                initialSaleCommission: artistInitial,
                secondaryMarketCommission: artistSecondary,
                address: artistAddress!
            )
        }
        if platformAddress != nil {
            roles["Platform"] = Evergreen.Role(
                id: "Platform",
                initialSaleCommission: platformInitial,
                secondaryMarketCommission: platformSecondary,
                address: platformAddress!
            )
        }

        self.admin.sealMaster(metadata: DigitalArt.Metadata(
            metadataLink: metadataLink,
            name: name,
            artist: artist,
            description: description,
            type: type,
            contentLink: contentLink,
            contentPreviewLink: contentPreviewLink,
            mimetype: mimetype,
            edition: 0,
            maxEdition: maxEdition,
            asset: asset,
            record: record,
            assetHead: assetHead,
            evergreenProfile: Evergreen.Profile(
                id: evergreenProfileID,
                roles: roles,
                description: ""
            )
        ))
    }
}
{{ end }}