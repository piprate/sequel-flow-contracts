{{ define "master_seal" }}
import NonFungibleToken from 0x{{.NFTAddress}}
import {{.TokenName}} from 0x{{.TokenAddress}}

transaction {
    let admin: &{{.TokenName}}.Admin

    prepare(signer: AuthAccount) {

        self.admin = signer.borrow<&{{.TokenName}}.Admin>(from: {{.AdminStoragePath}})!
    }

    execute {
        //let artistAddr: Address = 0x436164656E636521
        self.admin.sealMaster(metadata: {{.TokenName}}.Metadata(
            metadataLink: "{{.MD.MetadataLink}}",
            name: "{{.MD.Name}}",
            artist: "{{.MD.Artist}}",
            artistAddress: 0x{{.MD.ArtistAddress}},
            description: "{{.MD.Description}}",
            type: "{{.MD.Type}}",
            contentLink: "{{.MD.ContentLink}}",
            contentPreviewLink: "{{.MD.ContentPreviewLink}}",
            mimetype: "{{.MD.Mimetype}}",
            edition: 0,
            maxEdition: {{.MD.MaxEdition}},
            asset: "{{.MD.Asset}}",
            record: "{{.MD.Record}}",
            assetHead: "{{.MD.AssetHead}}",
        ))
    }
}
{{ end }}