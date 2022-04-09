import NonFungibleToken from "../../contracts/standard/NonFungibleToken.cdc"
import DigitalArt from "../../contracts/DigitalArt.cdc"

pub fun main(address:Address, tokenId:UInt64) : DigitalArt.Metadata? {
    let meta = DigitalArt.getMetadata(address: address, tokenId: tokenId)
    return meta
}
