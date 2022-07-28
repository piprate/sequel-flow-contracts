{{ define "account_switchboard_setup" }}
import FungibleTokenSwitchboard from {{.FungibleTokenSwitchboard}}
import FungibleToken from {{.FungibleToken}}
import MetadataViews from {{.MetadataViews}}

transaction {

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

    }
}
{{ end }}