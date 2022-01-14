{{ define "account_balance_fusd" }}
import FungibleToken from {{.FungibleToken}}
import FUSD from {{.FUSD}}

pub fun main(address: Address): UFix64 {
    let account = getAccount(address)

    let vaultRef = account.getCapability(/public/fusdBalance)!
        .borrow<&FUSD.Vault{FungibleToken.Balance}>()
        ?? panic("Could not borrow Balance reference to the Vault")

    return vaultRef.balance
}
{{ end }}