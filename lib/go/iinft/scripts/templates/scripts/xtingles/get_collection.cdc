{{ define "xtingles_get_collection" }}
import NonFungibleToken from 0x{{.NonFungibleToken}}
import Collectible from 0x{{.Collectible}}

pub fun main(address: Address): [XTinglesResource] {
    let account = getAccount(address)

    var collectibleData: [XTinglesResource] = []

    if let collectionRef = account.getCapability<&{Collectible.CollectionPublic}>(Collectible.CollectionPublicPath).borrow() {
        for id in collectionRef.getIDs() {
            var collectible = collectionRef.borrowCollectible(id: id)
            collectibleData.append(XTinglesResource(
                id: id,
                link: collectible!.metadata.link,
                name: collectible!.metadata.name,
                author: collectible!.metadata.author,
                description: collectible!.metadata.description,
                edition: collectible!.metadata.edition,
                editionNumber: collectible!.editionNumber
            ))
        }
    }

    return collectibleData
}

pub struct XTinglesResource {
    // The token's ID
    pub let id: UInt64
    // Link to IPFS file
    pub let link: String
    // Name
    pub let name: String
    // Author name
    pub let author: String
    // Description
    pub let description: String
    // Number of copy
    pub let edition: UInt64
    // Common number for all copies of the item
    pub let editionNumber: UInt64

    // initializer
    //
    init(id: UInt64, link: String, name: String, author: String, description: String, edition: UInt64, editionNumber: UInt64) {
        self.id = id
        self.link = link
        self.name = name
        self.author = author
        self.description = description
        self.edition = edition
        self.editionNumber = editionNumber
    }
}
{{ end }}