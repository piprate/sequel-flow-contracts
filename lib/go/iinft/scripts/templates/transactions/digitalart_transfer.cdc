{{ define "digitalart_transfer" }}
import NonFungibleToken from {{.NonFungibleToken}}
import DigitalArt from {{.DigitalArt}}

transaction(tokenId: UInt64, recipientAddr: Address) {
  prepare(acct: AuthAccount) {
    let recipient = getAccount(recipientAddr)
    let collectionRef = acct.borrow<&DigitalArt.Collection>(from: DigitalArt.CollectionStoragePath)!
    let depositRef = recipient.getCapability(DigitalArt.CollectionPublicPath)!.borrow<&{NonFungibleToken.CollectionPublic}>()!
    let nft <- collectionRef.withdraw(withdrawID: tokenId)
    depositRef.deposit(token: <-nft)
  }
}
{{ end }}