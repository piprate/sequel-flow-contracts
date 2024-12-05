{{ define "account_setup" }}
import NonFungibleToken from {{.NonFungibleToken}}
import MetadataViews from {{.MetadataViews}}
import Evergreen from {{.Evergreen}}
import DigitalArt from {{.DigitalArt}}

// This transaction is what an account would run
// to set itself up to receive NFTs

transaction {

    prepare(signer: auth(BorrowValue, IssueStorageCapabilityController, PublishCapability, SaveValue, UnpublishCapability) &Account) {

        let collectionData = DigitalArt.resolveContractView(resourceType: nil, viewType: Type<MetadataViews.NFTCollectionData>()) as! MetadataViews.NFTCollectionData?
            ?? panic("Could not resolve NFTCollectionData view. The DigitalArt contract needs to implement the NFTCollectionData Metadata view in order to execute this transaction")

        // Return early if the account already has a collection
        if signer.storage.borrow<&DigitalArt.Collection>(from: collectionData.storagePath) != nil {
            return
        }

        // Create a new empty collection
        let collection <- DigitalArt.createEmptyCollection(nftType: Type<@DigitalArt.NFT>())

        // save it to the account
        signer.storage.save(<-collection, to: collectionData.storagePath)

        // create a public capability for the collection
        signer.capabilities.unpublish(collectionData.publicPath)
        let collectionCap = signer.capabilities.storage.issue<&DigitalArt.Collection>(collectionData.storagePath)
        signer.capabilities.publish(collectionCap, at: collectionData.publicPath)
    }
}
{{ end }}
