{{ define "master_seal" }}
import NonFungibleToken from 0x{{.NFTAddress}}
import DigitalArt from 0x{{.TokenAddress}}

transaction(metadataLink: String,
            name: String,
            artist: String,
            artistAddress: Address,
            description: String,
            type: String,
            contentLink: String,
            contentPreviewLink: String,
            mimetype: String,
            maxEdition: UInt64,
            asset: String,
            record: String,
            assetHead: String) {
    let admin: &DigitalArt.Admin

    prepare(signer: AuthAccount) {

        self.admin = signer.borrow<&DigitalArt.Admin>(from: DigitalArt.AdminStoragePath)!
    }

    execute {
        self.admin.sealMaster(metadata: DigitalArt.Metadata(
            metadataLink: metadataLink,
            name: name,
            artist: artist,
            artistAddress: artistAddress,
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
        ))
    }
}
{{ end }}