{{ define "catalog_get_token_metadata" }}
import MetadataViews from {{.MetadataViews}}
import NFTCatalog from {{.NFTCatalog}}
import NFTRetrieval from {{.NFTRetrieval}}

pub struct NFT {
    pub let id: UInt64
    pub let name: String
    pub let description: String
    pub let thumbnail: String
    pub let externalURL: String

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

pub fun main(ownerAddress: Address, collectionIdentifier: String, id: UInt64): NFT? {
    let account = getAccount(ownerAddress)

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
        return nil
    }

    let views = NFTRetrieval.getNFTViewsFromIDs(collectionIdentifier: collectionIdentifier, ids: [id],  collectionCap: collectionCap)

    for view in views {
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