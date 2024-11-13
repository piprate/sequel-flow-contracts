{{ define "account_setup_flow_token" }}
import FungibleToken from {{.FungibleToken}}
import FlowToken from {{.FlowToken}}

transaction {

    prepare(signer: auth(BorrowValue, SaveValue, IssueStorageCapabilityController, PublishCapability) &Account) {
        if signer.storage.borrow<&FlowToken.Vault>(from: /storage/flowTokenVault) == nil {
            // Create a new flowToken Vault and put it in storage
            signer.storage.save(<-FlowToken.createEmptyVault(vaultType: Type<@FlowToken.Vault>()), to: /storage/flowTokenVault)

            // Create a public capability to the Vault that only exposes
            // the deposit function through the Receiver interface
            let vaultCap = signer.capabilities.storage.issue<&FlowToken.Vault>(/storage/flowTokenVault)
            signer.capabilities.publish(vaultCap, at: /public/flowTokenReceiver)

            // Create a public capability to the Vault that only exposes
            // the balance field through the Balance interface
            let balanceCap = signer.capabilities.storage.issue<&FlowToken.Vault>(/storage/flowTokenVault)
            signer.capabilities.publish(balanceCap, at: /public/flowTokenBalance)
        }
    }
}
{{ end }}
