{{ define "account_balance_flow" }}
// This script reads the balance field of an account's FlowToken Balance
import FungibleToken from {{.FungibleToken}}
import FlowToken from {{.FlowToken}}

pub fun main(account: Address): UFix64 {

    let vaultRef = getAccount(account)
        .getCapability(/public/flowTokenBalance)
        .borrow<&FlowToken.Vault{FungibleToken.Balance}>()
        ?? panic("Could not borrow Balance reference to the Vault")

    return vaultRef.balance
}
{{ end }}