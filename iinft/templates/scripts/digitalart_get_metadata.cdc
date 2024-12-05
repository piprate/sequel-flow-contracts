{{ define "digitalart_get_metadata" }}
import NonFungibleToken from {{.NonFungibleToken}}
import DigitalArt from {{.DigitalArt}}

access(all) fun main(address:Address, tokenId:UInt64) : &DigitalArt.Metadata? {
    let meta = DigitalArt.getMetadata(address: address, tokenId: tokenId)
    return meta

}
{{ end }}
