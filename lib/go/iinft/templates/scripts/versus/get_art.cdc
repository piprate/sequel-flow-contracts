{{ define "versus_get_art" }}
import Art from {{.Art}}

pub fun main(address: Address, artId: UInt64) : { String: String? } {
    let account = getAccount(address)

    let res : { String: String? } = {}
    if let artCollection= account.getCapability(Art.CollectionPublicPath).borrow<&{Art.CollectionPublic}>()  {
        let art = artCollection.borrowArt(id: artId)!
        res[art.cacheKey()] = art.content()
    }
    return res
}
{{ end }}