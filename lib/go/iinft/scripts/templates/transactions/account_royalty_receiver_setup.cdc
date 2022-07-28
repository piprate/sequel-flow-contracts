{{ define "account_royalty_receiver_setup" }}
import FungibleTokenSwitchboard from {{.FungibleTokenSwitchboard}}
import FungibleToken from {{.FungibleToken}}
import FlowToken from {{.FlowToken}}
import FUSD from {{.FUSD}}
import MetadataViews from {{.MetadataViews}}

transaction {

    let flowTokenVaultCapabilty: Capability<&{FungibleToken.Receiver}>
    let fusdTokenVaultCapabilty: Capability<&{FungibleToken.Receiver}>
    let switchboardRef:  &FungibleTokenSwitchboard.Switchboard

    prepare(acct: AuthAccount) {
        // Check if the account already has a Switchboard resource
        if acct.borrow<&FungibleTokenSwitchboard.Switchboard>
          (from: FungibleTokenSwitchboard.StoragePath) == nil {

            // Create a new Switchboard resource and put it into storage
            acct.save(
                <- FungibleTokenSwitchboard.createSwitchboard(),
                to: FungibleTokenSwitchboard.StoragePath)

            // Create a public capability to the Switchboard exposing the deposit
            // function through the {FungibleToken.Receiver} interface
            acct.link<&FungibleTokenSwitchboard.Switchboard{FungibleToken.Receiver}>(
                FungibleTokenSwitchboard.ReceiverPublicPath,
                target: FungibleTokenSwitchboard.StoragePath
            )

            // Create a public capability to the Switchboard exposing both the
            // deposit function and the getVaultCapabilities function through the
            // {FungibleTokenSwitchboard.SwitchboardPublic} interface
            acct.link<&FungibleTokenSwitchboard.Switchboard{FungibleTokenSwitchboard.SwitchboardPublic}>(
                FungibleTokenSwitchboard.PublicPath,
                target: FungibleTokenSwitchboard.StoragePath
            )
        }

        // Get a reference to the signers switchboard
        self.switchboardRef = acct.borrow<&FungibleTokenSwitchboard.Switchboard>
            (from: FungibleTokenSwitchboard.StoragePath)
            ?? panic("Could not borrow reference to switchboard")

        if acct.borrow<&FlowToken.Vault>(from: /storage/flowTokenVault) == nil {
            // Create a new flowToken Vault and put it in storage
            acct.save(<-FlowToken.createEmptyVault(), to: /storage/flowTokenVault)

            // Create a public capability to the Vault that only exposes
            // the deposit function through the Receiver interface
            acct.link<&FlowToken.Vault{FungibleToken.Receiver}>(
                /public/flowTokenReceiver,
                target: /storage/flowTokenVault
            )

            // Create a public capability to the Vault that only exposes
            // the balance field through the Balance interface
            acct.link<&FlowToken.Vault{FungibleToken.Balance}>(
                /public/flowTokenBalance,
                target: /storage/flowTokenVault
            )
        }

        self.flowTokenVaultCapabilty =
            acct.getCapability<&{FungibleToken.Receiver}>(/public/flowTokenReceiver)

        if(acct.borrow<&FUSD.Vault>(from: /storage/fusdVault) == nil) {
            // Create a new FUSD Vault and put it in storage
            acct.save(<-FUSD.createEmptyVault(), to: /storage/fusdVault)

            // Create a public capability to the Vault that only exposes
            // the deposit function through the Receiver interface
            acct.link<&FUSD.Vault{FungibleToken.Receiver}>(
                /public/fusdReceiver,
                target: /storage/fusdVault
            )

            // Create a public capability to the Vault that only exposes
            // the balance field through the Balance interface
            acct.link<&FUSD.Vault{FungibleToken.Balance}>(
                /public/fusdBalance,
                target: /storage/fusdVault
            )
        }

        self.fusdTokenVaultCapabilty =
            acct.getCapability<&{FungibleToken.Receiver}>(/public/fusdReceiver)
    }

    execute {
      self.switchboardRef.addNewVault(capability: self.flowTokenVaultCapabilty)
      self.switchboardRef.addNewVault(capability: self.fusdTokenVaultCapabilty)
    }
}
{{ end }}