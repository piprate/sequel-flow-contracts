{{ define "master_seal" }}
import NonFungibleToken from {{.NonFungibleToken}}
import Evergreen from {{.Evergreen}}
import DigitalArt from {{.DigitalArt}}

transaction(metadata: DigitalArt.Metadata, evergreenProfile: Evergreen.Profile) {
    let admin: &DigitalArt.Admin

    prepare(signer: AuthAccount) {
        self.admin = signer.borrow<&DigitalArt.Admin>(from: DigitalArt.AdminStoragePath)!
    }

    execute {
        self.admin.sealMaster(metadata: metadata, evergreenProfile: evergreenProfile)
    }
}
{{ end }}