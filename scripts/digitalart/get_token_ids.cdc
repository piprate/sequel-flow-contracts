import NonFungibleToken from "../../contracts/standard/NonFungibleToken.cdc"
import DigitalArt from "../../contracts/DigitalArt.cdc"

pub fun main(address:Address) : [UInt64] {
    let collection = getAccount(address)
        .getCapability(DigitalArt.CollectionPublicPath)
        .borrow<&{DigitalArt.CollectionPublic}>()
        ?? panic("Could not borrow DigitalArt collection")
    return collection.getIDs()
}
