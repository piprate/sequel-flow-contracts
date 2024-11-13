{{ define "master_seal" }}
import NonFungibleToken from {{.NonFungibleToken}}
import Evergreen from {{.Evergreen}}
import DigitalArt from {{.DigitalArt}}

transaction(metadata: DigitalArt.Metadata, evergreenProfile: Evergreen.Profile) {
    let admin: &DigitalArt.Admin

    prepare(signer: auth(BorrowValue) &Account) {
        self.admin = signer.storage.borrow<&DigitalArt.Admin>(from: DigitalArt.AdminStoragePath)!
    }

    execute {
        self.admin.sealMaster(metadata: metadata, evergreenProfile: evergreenProfile)
    }
}
{{ end }}
