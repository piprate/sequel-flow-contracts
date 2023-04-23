{{ define "catalog_get_collections" }}
import MetadataViews from {{.MetadataViews}}
import NFTCatalog from {{.NFTCatalog}}
import NFTRetrieval from {{.NFTRetrieval}}

pub struct NFTCollection {
    pub let id: String
    pub let name: String
    pub let description: String
    pub let squareImage: String
    pub let bannerImage: String
    pub let externalURL: String
    pub let count: Number

    init(
        id: String,
        name: String,
        description: String,
        squareImage: String,
        bannerImage: String,
        externalURL: String,
        count: Number
    ) {
        self.id = id
        self.name = name
        self.description = description
        self.squareImage = squareImage
        self.bannerImage = bannerImage
        self.externalURL = externalURL
        self.count = count
    }
}

pub fun main(ownerAddress: Address): {String: NFTCollection} {
    let account = getAuthAccount(ownerAddress)
    let collections: {String: NFTCollection} = {}

    NFTCatalog.forEachCatalogKey(fun (collectionIdentifier: String):Bool {
        let value = NFTCatalog.getCatalogEntry(collectionIdentifier: collectionIdentifier)!
        let keyHash = String.encodeHex(HashAlgorithm.SHA3_256.hash(collectionIdentifier.utf8))
        let tempPathStr = "catalog".concat(keyHash)
        let tempPublicPath = PublicPath(identifier: tempPathStr)!

        account.link<&{MetadataViews.ResolverCollection}>(
            tempPublicPath,
            target: value.collectionData.storagePath
        )

        let collectionCap = account.getCapability<&AnyResource{MetadataViews.ResolverCollection}>(tempPublicPath)

        if !collectionCap.check() {
            return true
        }

        let count = NFTRetrieval.getNFTCountFromCap(collectionIdentifier: collectionIdentifier, collectionCap: collectionCap)

        if count != 0 {
            collections[collectionIdentifier] = NFTCollection(
                id: collectionIdentifier,
                name: value.collectionDisplay.name,
                description: value.collectionDisplay.description,
                squareImage: value.collectionDisplay.squareImage.file.uri(),
                bannerImage: value.collectionDisplay.bannerImage.file.uri(),
                externalURL: value.collectionDisplay.externalURL.url,
                count: count
            )
        }
        return true
    })

    return collections
}
{{ end }}