import FlowToken from "../contracts/standard/FlowToken.cdc"

access(all) fun main(account: Address): UFix64 {

    let vaultRef = getAccount(account)
        .capabilities.borrow<&FlowToken.Vault>(/public/flowTokenBalance)
        ?? panic("Could not borrow Balance reference to the Vault")

    return vaultRef.balance
}
