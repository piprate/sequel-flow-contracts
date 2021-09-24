{{ define "account_setup" }}
import NonFungibleToken from 0x{{.NFTAddress}}
import {{.TokenName}} from 0x{{.TokenAddress}}

// This transaction is what an account would run
// to set itself up to receive NFTs

transaction {

    prepare(acct: AuthAccount) {

        // Return early if the account already has a collection
        if acct.borrow<&{{.TokenName}}.Collection>(from: {{.PrivateStoragePath}}) != nil {
            return
        }

        // Create a new empty collection
        let collection <- {{.TokenName}}.createEmptyCollection()

        // save it to the account
        acct.save(<-collection, to: {{.PrivateStoragePath}})

        // create a public capability for the collection
        acct.link<&{NonFungibleToken.CollectionPublic, {{.TokenName}}.CollectionPublic}>(
            {{.PublicStoragePath}},
            target: {{.PrivateStoragePath}}
        )
    }
}
{{ end }}