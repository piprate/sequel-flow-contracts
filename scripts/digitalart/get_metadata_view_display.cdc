import NonFungibleToken from "../../contracts/standard/NonFungibleToken.cdc"
import MetadataViews from "../../contracts/standard/MetadataViews.cdc"
import DigitalArt from "../../contracts/DigitalArt.cdc"

pub fun main(address:Address, tokenID:UInt64) : MetadataViews.Display? {
    let collection = getAccount(address).getCapability(DigitalArt.CollectionPublicPath)!.borrow<&{DigitalArt.CollectionPublic}>()!
    if let item = collection.borrowDigitalArt(id: tokenID) {
        if let view = item.resolveView(Type<MetadataViews.Display>()) {
            return view as! MetadataViews.Display
        }
    }

    return nil
}
