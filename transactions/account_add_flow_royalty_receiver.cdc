import FungibleTokenSwitchboard from "../contracts/standard/FungibleTokenSwitchboard.cdc"
import FungibleToken from "../contracts/standard/FungibleToken.cdc"
import FungibleTokenMetadataViews from "../contracts/standard/FungibleTokenMetadataViews.cdc"
import FlowToken from "../contracts/standard/FlowToken.cdc"
import MetadataViews from "../contracts/standard/MetadataViews.cdc"

transaction {

    let tokenVaultCapabilities: [Capability<&{FungibleToken.Receiver}>]
    let switchboardRef:  auth(FungibleTokenSwitchboard.Owner) &FungibleTokenSwitchboard.Switchboard

    prepare(signer: auth(BorrowValue, SaveValue, IssueStorageCapabilityController, PublishCapability, UnpublishCapability) &Account) {
        // Check if the account already has a Switchboard resource
        if signer.storage.borrow<&FungibleTokenSwitchboard.Switchboard>(from: FungibleTokenSwitchboard.StoragePath) == nil {
            // Create a new Switchboard and save it in storage
            signer.storage.save(<-FungibleTokenSwitchboard.createSwitchboard(), to: FungibleTokenSwitchboard.StoragePath)
            // Clear existing Capabilities at canonical paths
            signer.capabilities.unpublish(FungibleTokenSwitchboard.ReceiverPublicPath)
            signer.capabilities.unpublish(FungibleTokenSwitchboard.PublicPath)
            // Issue Receiver & Switchboard Capabilities
            let receiverCap = signer.capabilities.storage.issue<&{FungibleToken.Receiver}>(
                    FungibleTokenSwitchboard.StoragePath
                )
            let switchboardPublicCap = signer.capabilities.storage.issue<&{FungibleTokenSwitchboard.SwitchboardPublic, FungibleToken.Receiver}>(
                    FungibleTokenSwitchboard.StoragePath
                )
            // Publish Capabilities
            signer.capabilities.publish(receiverCap, at: FungibleTokenSwitchboard.ReceiverPublicPath)
            signer.capabilities.publish(switchboardPublicCap, at: FungibleTokenSwitchboard.PublicPath)
        }

        // Get a reference to the account's switchboard
        self.switchboardRef = signer.storage.borrow<auth(FungibleTokenSwitchboard.Owner) &FungibleTokenSwitchboard.Switchboard>(
                from: FungibleTokenSwitchboard.StoragePath)
            ?? panic("Could not borrow reference to switchboard")

        if signer.storage.borrow<&FlowToken.Vault>(from: /storage/flowTokenVault) == nil {
            // Create a new flowToken Vault and put it in storage
            signer.storage.save(<-FlowToken.createEmptyVault(vaultType: Type<@FlowToken.Vault>()), to: /storage/flowTokenVault)

            // Create a public capability to the Vault that only exposes
            // the deposit function through the Receiver interface
            let vaultCap = signer.capabilities.storage.issue<&FlowToken.Vault>(
                /storage/flowTokenVault
            )

            signer.capabilities.publish(
                vaultCap,
                at: /public/flowTokenReceiver
            )

            // Create a public capability to the Vault that only exposes
            // the balance field through the Balance interface
            let balanceCap = signer.capabilities.storage.issue<&FlowToken.Vault>(
                /storage/flowTokenVault
            )

            signer.capabilities.publish(
                balanceCap,
                at: /public/flowTokenBalance
            )
        }

        self.tokenVaultCapabilities = [signer.capabilities.get<&{FungibleToken.Receiver}>(/public/flowTokenReceiver)]
    }

    execute {
        for cap in self.tokenVaultCapabilities {
            self.switchboardRef.addNewVault(capability: cap)
        }

    }
}
