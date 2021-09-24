{{ define "token_mint" }}
import NonFungibleToken from 0x{{.NFTAddress}}
import {{.TokenName}} from 0x{{.TokenAddress}}

transaction {
    let admin: &{{.TokenName}}.Admin

    prepare(signer: AuthAccount) {

        self.admin = signer.borrow<&{{.TokenName}}.Admin>(from: {{.AdminStoragePath}})!
    }

    execute {
        let recipient = getAccount(0x{{.ReceiverAddress}})

        let receiver = recipient
            .getCapability({{.ReceiverPublicCollection}})!
            .borrow<&{ {{.TokenName}}.CollectionPublic}>()
            ?? panic("Could not get receiver reference to the NFT Collection")

        let newNFT <- self.admin.mintNFT(masterId: "{{.Master}}")

        receiver.deposit(token: <-newNFT)
    }
}
{{ end }}