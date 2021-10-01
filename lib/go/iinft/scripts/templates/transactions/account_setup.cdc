{{ define "account_setup" }}
import NonFungibleToken from 0x{{.NFTAddress}}
import DigitalArt from 0x{{.TokenAddress}}

// This transaction is what an account would run
// to set itself up to receive NFTs

transaction {

    prepare(acct: AuthAccount) {

        // Return early if the account already has a collection
        if acct.borrow<&DigitalArt.Collection>(from: DigitalArt.CollectionStoragePath) != nil {
            return
        }

        // Create a new empty collection
        let collection <- DigitalArt.createEmptyCollection()

        // save it to the account
        acct.save(<-collection, to: DigitalArt.CollectionStoragePath)

        // create a public capability for the collection
        acct.link<&{NonFungibleToken.CollectionPublic, DigitalArt.CollectionPublic}>(
            DigitalArt.CollectionPublicPath,
            target: DigitalArt.CollectionStoragePath
        )
    }
}
{{ end }}