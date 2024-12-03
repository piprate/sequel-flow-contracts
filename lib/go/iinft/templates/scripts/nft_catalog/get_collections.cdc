{{ define "catalog_get_collections" }}
import MetadataViews from {{.MetadataViews}}
import NFTCatalog from {{.NFTCatalog}}
import ViewResolver from {{.ViewResolver}}

access(all) struct NFTCollection {
    access(all) let id: String
    access(all) let name: String
    access(all) let description: String
    access(all) let squareImage: String
    access(all) let bannerImage: String
    access(all) let externalURL: String
    access(all) let count: Number

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

access(all) fun main(ownerAddress: Address): {String: NFTCollection} {
    let account = getAccount(ownerAddress)
    let collections: {String: NFTCollection} = {}

    fun hasMultipleCollectionsFn(nftTypeIdentifier : String): Bool {
        let typeCollections = NFTCatalog.getCollectionsForType(nftTypeIdentifier: nftTypeIdentifier)!
        var numberOfCollections = 0
        for identifier in typeCollections.keys {
            let existence = typeCollections[identifier]!
            if existence {
                numberOfCollections = numberOfCollections + 1
            }
            if numberOfCollections > 1 {
                return true
            }
        }
        return false
    }

    NFTCatalog.forEachCatalogKey(fun (collectionIdentifier: String):Bool {
        let value = NFTCatalog.getCatalogEntry(collectionIdentifier: collectionIdentifier)!

        let collectionCap = account.capabilities.get<&{ViewResolver.ResolverCollection}>(value.collectionData.publicPath)

        if !collectionCap.check() {
            return true
        }

        // Check if we have multiple collections for the NFT type...
        let hasMultipleCollections = hasMultipleCollectionsFn(nftTypeIdentifier : value.nftType.identifier)

        var count : UInt64 = 0
        let collectionRef = collectionCap.borrow()!
        if !hasMultipleCollections {
            count = UInt64(collectionRef.getIDs().length)
        } else {
            for id in collectionRef.getIDs() {
                let nftResolver = collectionRef.borrowViewResolver(id: id)
                let nftViews = MetadataViews.getNFTView(id: id, viewResolver: nftResolver!)
                if nftViews.display!.name == value.collectionDisplay.name {
                    count = count + 1
                }
            }
        }

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
