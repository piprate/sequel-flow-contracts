{{ define "versus_get_art" }}
import Art from {{.Art}}

pub fun main(address: Address, artId: UInt64) : String? {
    let account = getAccount(address)
    return Art.getContentForArt(address: address, artId: artId)
}
{{ end }}