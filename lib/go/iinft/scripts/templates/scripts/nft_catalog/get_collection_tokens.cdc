{{ define "catalog_get_collection_tokens" }}
import MetadataViews from {{.MetadataViews}}
import NFTCatalog from {{.NFTCatalog}}
import NFTRetrieval from {{.NFTRetrieval}}

pub struct NFT {
    pub let id: UInt64
    pub let name: String
    pub let description: String
    pub let thumbnail: String
    pub let externalURL: String
    pub let storagePath: StoragePath
    pub let publicPath: PublicPath
    pub let privatePath: PrivatePath
    pub let publicLinkedType: Type
    pub let privateLinkedType: Type
    pub let royalties: [MetadataViews.Royalty]

    init(
        id: UInt64,
        name: String,
        description: String,
        thumbnail: String,
        externalURL: String,
        storagePath: StoragePath,
        publicPath: PublicPath,
        privatePath: PrivatePath,
        publicLinkedType: Type,
        privateLinkedType: Type,
        royalties: [MetadataViews.Royalty]
    ) {
        self.id = id
        self.name = name
        self.description = description
        self.thumbnail = thumbnail
        self.externalURL = externalURL
        self.storagePath = storagePath
        self.publicPath = publicPath
        self.privatePath = privatePath
        self.publicLinkedType = publicLinkedType
        self.privateLinkedType = privateLinkedType
        self.royalties = royalties
    }
}

pub fun main(ownerAddress: Address, collectionIdentifier: String): [NFT] {
    let account = getAuthAccount(ownerAddress)

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
        return []
    }

    let views = NFTRetrieval.getNFTViewsFromCap(collectionIdentifier: collectionIdentifier, collectionCap: collectionCap)

    let items: [NFT] = []

    for view in views {
        let displayView = view.display
        let externalURLView = view.externalURL
        let collectionDataView = view.collectionData
        let collectionDisplayView = view.collectionDisplay
        let royaltyView = view.royalties

        if (displayView == nil || externalURLView == nil || royaltyView == nil) {
            // Bad NFT. Skipping....
            continue
        }

        items.append(
            NFT(
                id: view.id,
                name: displayView!.name,
                description: displayView!.description,
                thumbnail: displayView!.thumbnail.uri(),
                externalURL: externalURLView!.url,
                storagePath: collectionDataView!.storagePath,
                publicPath: collectionDataView!.publicPath,
                privatePath: collectionDataView!.providerPath,
                publicLinkedType: collectionDataView!.publicLinkedType,
                privateLinkedType: collectionDataView!.providerLinkedType,
                royalties: royaltyView!.getRoyalties()
            )
        )
    }

    return items
}
{{ end }}