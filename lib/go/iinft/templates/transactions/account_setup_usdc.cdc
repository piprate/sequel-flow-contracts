{{ define "account_setup_usdc" }}
import FungibleToken from {{.FungibleToken}}
import USDCFlow from {{.USDCFlow}}

/// This script is used to add a USDCFlow.Vault resource to the signer's account
/// so that they can use USDCFlow
///
/// If the Vault already exist for the account,
/// the script will return immediately without error
///

transaction {

    prepare(signer: auth(Storage, BorrowValue, Capabilities, AddContract) &Account) {

        // Return early if the account already stores a USDCFlow Vault
        if signer.storage.borrow<&USDCFlow.Vault>(from: USDCFlow.VaultStoragePath) != nil {
            return
        }

        // Create a new ExampleToken Vault and put it in storage
        signer.storage.save(
            <-USDCFlow.createEmptyVault(vaultType: Type<@USDCFlow.Vault>()),
            to: USDCFlow.VaultStoragePath
        )

        let receiver = signer.capabilities.storage.issue<&USDCFlow.Vault>(
            USDCFlow.VaultStoragePath
        )
        signer.capabilities.publish(receiver, at: USDCFlow.ReceiverPublicPath)
        signer.capabilities.publish(receiver, at: USDCFlow.VaultPublicPath)
    }
}
{{ end }}
