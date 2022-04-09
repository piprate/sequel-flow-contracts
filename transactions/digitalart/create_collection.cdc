import NonFungibleToken from "../../contracts/standard/NonFungibleToken.cdc"
import DigitalArt from "../../contracts/DigitalArt.cdc"

// This transaction is what an account would run
// to set itself up to receive Digital Art NFTs

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
        acct.link<&{NonFungibleToken.CollectionPublic, NonFungibleToken.Receiver, DigitalArt.CollectionPublic}>(
            DigitalArt.CollectionPublicPath,
            target: DigitalArt.CollectionStoragePath
        )
    }
}
