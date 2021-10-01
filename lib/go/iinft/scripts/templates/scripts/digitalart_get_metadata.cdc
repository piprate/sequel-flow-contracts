{{ define "digitalart_get_metadata" }}
import NonFungibleToken from 0x{{.NonFungibleToken}}
import DigitalArt from 0x{{.DigitalArt}}

pub fun main(address:Address, tokenId:UInt64) : DigitalArt.Metadata? {
    let meta = DigitalArt.getMetadata(address: address, tokenId: tokenId)
    return meta

}
{{ end }}