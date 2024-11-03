{{ define "digitalart_transfer" }}
import NonFungibleToken from {{.NonFungibleToken}}
import DigitalArt from {{.DigitalArt}}

transaction(tokenId: UInt64, recipientAddr: Address) {
  prepare(acct: auth(BorrowValue) &Account) {
    let recipient = getAccount(recipientAddr)
    let collectionRef = acct.storage.borrow<auth(NonFungibleToken.Withdraw) &DigitalArt.Collection>(from: DigitalArt.CollectionStoragePath)!
    let depositRef = recipient.capabilities.borrow<&{NonFungibleToken.Receiver}>(DigitalArt.CollectionPublicPath)!
    let nft <- collectionRef.withdraw(withdrawID: tokenId)
    depositRef.deposit(token: <-nft)
  }
}
{{ end }}
