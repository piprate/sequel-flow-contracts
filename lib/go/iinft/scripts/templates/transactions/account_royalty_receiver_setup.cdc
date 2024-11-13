{{ define "account_royalty_receiver_setup" }}
import FungibleTokenSwitchboard from {{.FungibleTokenSwitchboard}}
import FungibleToken from {{.FungibleToken}}
import FungibleTokenMetadataViews from {{.FungibleTokenMetadataViews}}
import FlowToken from {{.FlowToken}}
import MetadataViews from {{.MetadataViews}}

transaction(extraTokenContractAddresses: [Address], extraTokenContractNames: [String]) {

    let tokenVaultCapabilities: [Capability<&{FungibleToken.Receiver}>]
    let switchboardRef:  auth(FungibleTokenSwitchboard.Owner) &FungibleTokenSwitchboard.Switchboard

    prepare(signer: auth(BorrowValue, SaveValue, IssueStorageCapabilityController, PublishCapability, UnpublishCapability) &Account) {
        assert(extraTokenContractAddresses.length == extraTokenContractNames.length, message: "lengths of extraTokenContractAddresses and extraTokenContractNames should be equal")

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

        var i = 0
        for contractAddress in extraTokenContractAddresses {
            let contractName = extraTokenContractNames[i]

            // Borrow a reference to the vault stored on the passed account at the passed publicPath
            let resolverRef = getAccount(contractAddress)
                .contracts.borrow<&{FungibleToken}>(name: contractName)
                    ?? panic("Could not borrow FungibleToken reference to the contract. Make sure the provided contract name ("
                              .concat(contractName).concat(") and address (").concat(contractAddress.toString()).concat(") are correct!"))

            // Use that reference to retrieve the FTView
            let ftVaultData = resolverRef.resolveContractView(resourceType: nil, viewType: Type<FungibleTokenMetadataViews.FTVaultData>()) as! FungibleTokenMetadataViews.FTVaultData?
                ?? panic("Could not resolve FTVaultData view. The ".concat(contractName)
                    .concat(" contract needs to implement the FTVaultData Metadata view in order to execute this transaction."))

            if signer.storage.borrow<&{FungibleToken.Vault}>(from: ftVaultData.storagePath) == nil {
                // Create a new empty vault using the createEmptyVault function inside the FTVaultData
                let emptyVault <-ftVaultData.createEmptyVault()

                // Save it to the account
                signer.storage.save(<-emptyVault, to: ftVaultData.storagePath)

                // Create a public capability for the vault which includes the .Resolver interface
                let vaultCap = signer.capabilities.storage.issue<&{FungibleToken.Vault}>(ftVaultData.storagePath)
                signer.capabilities.publish(vaultCap, at: ftVaultData.metadataPath)

                // Create a public capability for the vault exposing the receiver interface
                let receiverCap = signer.capabilities.storage.issue<&{FungibleToken.Receiver}>(ftVaultData.storagePath)
                signer.capabilities.publish(receiverCap, at: ftVaultData.receiverPath)
            }

            let cap = signer.capabilities.get<&{FungibleToken.Receiver}>(ftVaultData.receiverPath)
            self.tokenVaultCapabilities.append(cap)

            i = i + 1
        }
    }

    execute {
        for cap in self.tokenVaultCapabilities {
            self.switchboardRef.addNewVault(capability: cap)
        }

    }
}
{{ end }}
