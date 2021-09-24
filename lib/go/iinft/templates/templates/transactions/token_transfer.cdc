{{ define "token_transfer" }}
import NonFungibleToken from 0x{{.NFTAddress}}
import {{.TokenName}} from 0x{{.TokenAddress}}

transaction {
  prepare(acct: AuthAccount) {
    let recipient = getAccount(0x{{.RecipientAddress}})
    let collectionRef = acct.borrow<&{{.TokenName}}.Collection>(from: {{.SenderStorageCollection}})!
    let depositRef = recipient.getCapability({{.RecipientPublicCollection}})!.borrow<&{NonFungibleToken.CollectionPublic}>()!
    let nft <- collectionRef.withdraw(withdrawID: {{.TokenID}})
    depositRef.deposit(token: <-nft)
  }
}
{{ end }}