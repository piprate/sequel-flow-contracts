{{ define "account_balance_example_ft" }}
// This script reads the balance field of an account's ExampleToken Balance
import FungibleToken from {{.FungibleToken}}
import ExampleToken from {{.ExampleToken}}

access(all) fun main(account: Address): UFix64 {

    let vaultRef = getAccount(account).capabilities.borrow<&{FungibleToken.Balance}>(/public/exampleTokenVault)

    if vaultRef !=  nil {
        return vaultRef!.balance
    } else {
        return 0.0
    }
}
{{ end }}
