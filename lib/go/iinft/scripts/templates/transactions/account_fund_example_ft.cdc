{{ define "account_fund_example_ft" }}
import FungibleToken from {{.FungibleToken}}
import ExampleToken from {{.ExampleToken}}
import FungibleTokenMetadataViews from {{.FungibleTokenMetadataViews}}

/// This transaction is what the minter Account uses to mint new tokens
/// They provide the recipient address and amount to mint, and the tokens
/// are transferred to the address after minting

transaction(recipient: Address, amount: UFix64) {

    let tokenMinter: &ExampleToken.Minter
    let tokenReceiver: &{FungibleToken.Receiver}

    /// The total supply of tokens before the burn
    let supplyBefore: UFix64

    prepare(signer: auth(BorrowValue) &Account) {
        self.supplyBefore = ExampleToken.totalSupply

        // Borrow a reference to the admin object
        self.tokenMinter = signer.storage.borrow<&ExampleToken.Minter>(from: ExampleToken.AdminStoragePath)
            ?? panic("Cannot mint: Signer does not store the ExampleToken Minter in their account!")

        let vaultData = ExampleToken.resolveContractView(resourceType: nil, viewType: Type<FungibleTokenMetadataViews.FTVaultData>()) as! FungibleTokenMetadataViews.FTVaultData?
            ?? panic("Could not resolve FTVaultData view. The ExampleToken"
                .concat(" contract needs to implement the FTVaultData Metadata view in order to execute this transaction."))

        self.tokenReceiver = getAccount(recipient).capabilities.borrow<&{FungibleToken.Receiver}>(vaultData.receiverPath)
            ?? panic("Could not borrow a Receiver reference to the FungibleToken Vault in account "
                .concat(recipient.toString()).concat(" at path ").concat(vaultData.receiverPath.toString())
                .concat(". Make sure you are sending to an address that has ")
                .concat("a FungibleToken Vault set up properly at the specified path."))
    }

    execute {
        let mintedVault <- self.tokenMinter.mintTokens(amount: amount)

        self.tokenReceiver.deposit(from: <-mintedVault)
    }
}
{{ end }}
