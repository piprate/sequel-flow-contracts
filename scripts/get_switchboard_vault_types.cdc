import FungibleTokenSwitchboard from "../contracts/standard/FungibleTokenSwitchboard.cdc"
import FungibleToken from "../contracts/standard/FungibleToken.cdc"

access(all) fun main(account: Address): [Type] {
    let acct = getAccount(account)
    // Get a reference to the switchboard conforming to FungibleToken.Receiver
    let switchboardRef = acct.capabilities.borrow<&{FungibleToken.Receiver}>(FungibleTokenSwitchboard.ReceiverPublicPath)

    if switchboardRef == nil {
      return []
    }

    return switchboardRef!.getSupportedVaultTypes().keys
}
