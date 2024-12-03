{{ define "catalog_get_collection_tokens" }}
import MetadataViews from {{.MetadataViews}}
import NFTCatalog from {{.NFTCatalog}}

access(all) struct NFT {
    access(all) let id: UInt64
    access(all) let name: String
    access(all) let description: String
    access(all) let thumbnail: String
    access(all) let externalURL: String

    init(
        id: UInt64,
        name: String,
        description: String,
        thumbnail: String,
        externalURL: String
    ) {
        self.id = id
        self.name = name
        self.description = description
        self.thumbnail = thumbnail
        self.externalURL = externalURL
    }
}

access(all) fun main(ownerAddress: Address, collectionIdentifier: String): [NFT] {
    let account = getAuthAccount(ownerAddress)

    let value = NFTCatalog.getCatalogEntry(collectionIdentifier: collectionIdentifier)!
    let keyHash = String.encodeHex(HashAlgorithm.SHA3_256.hash(collectionIdentifier.utf8))
    let tempPathStr = "catalog".concat(keyHash)
    let tempPublicPath = PublicPath(identifier: tempPathStr)!

    let collectionCap = account.getCapability<&AnyResource{MetadataViews.ResolverCollection}>(tempPublicPath)

    if !collectionCap.check() {
        return []
    }

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

    let views : [MetadataViews.NFTView] = []

    // Check if we have multiple collections for the NFT type...
    let hasMultipleCollections = self.hasMultipleCollections(nftTypeIdentifier : value.nftType.identifier)

    if collectionCap.check() {
        let collectionRef = collectionCap.borrow()!
        for id in collectionRef.getIDs() {
            let nftResolver = collectionRef.borrowViewResolver(id: id)
            let nftViews = MetadataViews.getNFTView(id: id, viewResolver: nftResolver!)
            if !hasMultipleCollections {
                views.append(nftViews)
            } else if nftViews.display!.name == value.collectionDisplay.name {
                views.append(nftViews)
            }

        }
    }

    let items: [NFT] = []

    for view in views {
        let displayView = view.display
        let externalURLView = view.externalURL

        if (displayView == nil || externalURLView == nil) {
            // Bad NFT. Skipping....
            continue
        }

        items.append(
            NFT(
                id: view.id,
                name: displayView!.name,
                description: displayView!.description,
                thumbnail: displayView!.thumbnail.uri(),
                externalURL: externalURLView!.url
            )
        )
    }

    return items
}
{{ end }}
