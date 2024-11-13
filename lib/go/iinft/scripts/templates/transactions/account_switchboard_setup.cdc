{{ define "account_switchboard_setup" }}
import FungibleTokenSwitchboard from {{.FungibleTokenSwitchboard}}
import FungibleToken from {{.FungibleToken}}
import MetadataViews from {{.MetadataViews}}

transaction {

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
    }
}
{{ end }}
