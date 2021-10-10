{{ define "xtingles_get_collection" }}
import NonFungibleToken from 0x{{.NonFungibleToken}}
import Collectible from 0x{{.Collectible}}

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