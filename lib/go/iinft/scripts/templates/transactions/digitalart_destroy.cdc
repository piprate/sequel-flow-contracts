{{ define "digitalart_destroy" }}
import NonFungibleToken from 0x{{.NonFungibleToken}}
import DigitalArt from 0x{{.DigitalArt}}

transaction(tokenId: UInt64) {
  prepare(acct: AuthAccount) {
    let collection <- acct.load<@DigitalArt.Collection>(from: DigitalArt.CollectionStoragePath)!
    let nft <- collection.withdraw(withdrawID: tokenId)
    destroy nft

    acct.save(<-collection, to: DigitalArt.CollectionStoragePath)
  }
}
{{ end }}