{{ define "xtingles_get_collection" }}
import NonFungibleToken from {{.NonFungibleToken}}
import Collectible from {{.Collectible}}

pub fun main(address: Address): [UInt64] {
    let account = getAccount(address)

    if let collectionRef = account.getCapability<&{Collectible.CollectionPublic}>(Collectible.CollectionPublicPath).borrow() {
        return collectionRef.getIDs()
    } else {
        var emptyCollection: [UInt64] = []
        return emptyCollection
    }
}
{{ end }}