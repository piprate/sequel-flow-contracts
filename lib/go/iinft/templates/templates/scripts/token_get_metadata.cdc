{{ define "token_get_metadata" }}
import NonFungibleToken from 0x{{.NFTAddress}}
import {{.TokenName}} from 0x{{.TokenAddress}}

pub fun main(address:Address, tokenId:UInt64) : {{.TokenName}}.Metadata? {
    let meta = {{.TokenName}}.getMetadata(address: address, tokenId: tokenId)
    return meta

}
{{ end }}