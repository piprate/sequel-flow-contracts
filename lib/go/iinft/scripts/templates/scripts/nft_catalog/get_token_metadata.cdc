{{ define "catalog_get_token_metadata" }}
import MetadataViews from {{.MetadataViews}}
import NFTCatalog from {{.NFTCatalog}}
import ViewResolver from {{.ViewResolver}}

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

access(all) fun main(ownerAddress: Address, collectionIdentifier: String, id: UInt64): NFT? {
    pre {
         NFTCatalog.getCatalogEntry(collectionIdentifier: collectionIdentifier) != nil : "Invalid collection identifier"
    }

    let account = getAccount(ownerAddress)

    let value = NFTCatalog.getCatalogEntry(collectionIdentifier: collectionIdentifier)!

    let collectionCap = account.capabilities.get<&{ViewResolver.ResolverCollection}>(value.collectionData.publicPath)
    if !collectionCap.check() {
        return nil
    }

    let items : [MetadataViews.NFTView] = []

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

    // Check if we have multiple collections for the NFT type...
    let hasMultipleCollections = hasMultipleCollectionsFn(nftTypeIdentifier : value.nftType.identifier)

     if collectionCap.check() {
        let collectionRef = collectionCap.borrow()!
        if collectionRef.getIDs().contains(id) {
            let nftResolver = collectionRef.borrowViewResolver(id: id)
            let nftViews = MetadataViews.getNFTView(id: id, viewResolver: nftResolver!)
            if !hasMultipleCollections {
                items.append(nftViews)
            } else if nftViews.display!.name == value.collectionDisplay.name {
                items.append(nftViews)
            }
        }
    }

    for view in items {
        let displayView = view.display
        let externalURLView = view.externalURL

        if (view.id != id) {
            continue
        }

        if (displayView == nil || externalURLView == nil) {
            // Bad NFT. Skipping....
            continue
        }

        return NFT(
            id: view.id,
            name: displayView!.name,
            description: displayView!.description,
            thumbnail: displayView!.thumbnail.uri(),
            externalURL: externalURLView!.url
        )
    }

    return nil
}
{{ end }}
