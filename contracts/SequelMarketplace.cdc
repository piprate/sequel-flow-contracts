import FungibleToken from "./standard/FungibleToken.cdc"
import NFTStorefront from "./standard/NFTStorefront.cdc"
import NonFungibleToken from "./standard/NonFungibleToken.cdc"
import Evergreen from "./Evergreen.cdc"

// SequelMarketplace provides convenience functions to create listings for Sequel NFTs in NFTStorefront.
//
pub contract SequelMarketplace {
    // Payment
    //
    pub struct Payment {
        pub let role: String
        pub let receiver: Address
        pub let amount: UFix64
        pub let rate: UFix64

        init(role: String, receiver: Address, amount: UFix64, rate: UFix64) {
            self.role = role
            self.receiver = receiver
            self.amount = amount
            self.rate = rate
        }
    }

    // TokenListed
    // Token available for purchase
    //
    pub event TokenListed(
        storefrontAddress: Address,
        listingID: UInt64,
        nftType: String,
        nftID: UInt64,
        paymentVaultType: String,
        price: UFix64,
        payments: [Payment],
        asset: String,
        metadataLink: String?,
    )

    pub event TokenSold(
        storefrontAddress: Address,
        listingID: UInt64,
        nftType: String,
        nftID: UInt64,
        paymentVaultType: String,
        price: UFix64,
        buyerAddress: Address,
        metadataLink: String?,
    )

    pub event TokenWithdrawn(
        storefrontAddress: Address,
        listingID: UInt64,
        nftType: String,
        nftID: UInt64,
        vaultType: String,
        price: UFix64
    )

    // listToken
    pub fun listToken(
        storefront: &NFTStorefront.Storefront,
        nftProviderCapability: Capability<&{NonFungibleToken.Provider, NonFungibleToken.CollectionPublic, Evergreen.CollectionPublic}>,
        nftType: Type,
        nftID: UInt64,
        paymentVaultPath: PublicPath,
        paymentVaultType: Type,
        price: UFix64,
        extraRoles: [Evergreen.Role],
        metadataLink: String?,
    ): UInt64 {
        let token = nftProviderCapability.borrow()!.borrowEvergreenToken(id: nftID)!
        let seller = storefront.owner!.address

        let payments = self.buildPayments(
            profile: token.getEvergreenProfile(),
            seller: seller,
            sellerRole: "Owner",
            price: price,
            initialSale: false,
            extraRoles: extraRoles)

        let saleCuts: [NFTStorefront.SaleCut] = []
        for payment in payments {
            let receiver = getAccount(payment.receiver).getCapability<&{FungibleToken.Receiver}>(paymentVaultPath)
            assert(receiver.borrow() != nil, message: "Missing or mis-typed fungible token receiver")

            saleCuts.append(NFTStorefront.SaleCut(receiver: receiver, amount: payment.amount))
        }

        let listingID = storefront.createListing(
            nftProviderCapability: nftProviderCapability,
            nftType: nftType,
            nftID: nftID,
            salePaymentVaultType: paymentVaultType,
            saleCuts: saleCuts
        )

        emit TokenListed(
            storefrontAddress: storefront.owner!.address,
            listingID: listingID,
            nftType: nftType.identifier,
            nftID: nftID,
            paymentVaultType: paymentVaultType.identifier,
            price: price,
            payments: payments,
            asset: token.getAssetID(),
            metadataLink: metadataLink,
        )

        return listingID
    }

    pub fun buyToken(
        storefrontAddress: Address,
        storefront: &NFTStorefront.Storefront{NFTStorefront.StorefrontPublic},
        listingID: UInt64,
        listing: &NFTStorefront.Listing{NFTStorefront.ListingPublic},
        paymentVault: @FungibleToken.Vault,
        buyerAddress: Address,
        metadataLink: String?,
    ): @NonFungibleToken.NFT {
        let details = listing.getDetails()

        emit TokenSold(
            storefrontAddress: storefrontAddress,
            listingID: listingID,
            nftType: details.nftType.identifier,
            nftID: details.nftID,
            paymentVaultType: details.salePaymentVaultType.identifier,
            price: details.salePrice,
            buyerAddress: buyerAddress,
            metadataLink: metadataLink
        )

        let item <- listing.purchase(payment: <-paymentVault)
        storefront.cleanup(listingResourceID: listingID)
        return <- item
    }

    pub fun payForMintedTokens(
        unitPrice: UFix64,
        numEditions: UInt64,
        paymentVaultPath: PublicPath,
        paymentVault: @FungibleToken.Vault,
        evergreenProfile: Evergreen.Profile,
    ) {
        let artistAddress = evergreenProfile.roles["Artist"]!.address

        let payments = self.buildPayments(
            profile: evergreenProfile,
            seller: artistAddress,
            sellerRole: "Artist",
            price: unitPrice * UFix64(numEditions),
            initialSale: true,
            extraRoles: [])

        // Rather than aborting the transaction if any receiver is absent when we try to pay it,
        // we send the cut to the first valid receiver.
        // The first receiver should therefore either be the seller, or an agreed recipient for
        // any unpaid cuts.
        var residualReceiver: &{FungibleToken.Receiver}? = nil

        for payment in payments {
            let receiverCap = getAccount(payment.receiver).getCapability<&{FungibleToken.Receiver}>(paymentVaultPath)
            let receiver = receiverCap.borrow() ?? panic("Missing or mis-typed fungible token receiver")

            let paymentCut <- paymentVault.withdraw(amount: payment.amount)
            receiver.deposit(from: <-paymentCut)
            if (residualReceiver == nil) {
                residualReceiver = receiver
            }
        }

        // At this point, if all recievers were active and availabile, then the payment Vault will have
        // zero tokens left.
        if paymentVault.balance > 0.0 {
            assert(residualReceiver != nil, message: "No valid residual payment receivers")
            residualReceiver!.deposit(from: <-paymentVault)
        } else {
            destroy paymentVault
        }
    }

    // withdrawToken
    // Cancel sale
    //
    pub fun withdrawToken(
        storefrontAddress: Address,
        storefront: &NFTStorefront.Storefront,
        listingID: UInt64,
        listing: &NFTStorefront.Listing{NFTStorefront.ListingPublic},
    ) {
        let details = listing.getDetails()

        emit TokenWithdrawn(
            storefrontAddress: storefrontAddress,
            listingID: listingID,
            nftType: details.nftType.identifier,
            nftID: details.nftID,
            vaultType: details.salePaymentVaultType.identifier,
            price: details.salePrice
        )

        storefront.removeListing(listingResourceID: listingID)
    }

    pub fun buildPayments(
        profile: Evergreen.Profile,
        seller: Address,
        sellerRole: String,
        price: UFix64,
        initialSale: Bool,
        extraRoles: [Evergreen.Role]
    ): [Payment] {

        let payments: [Payment] = []
        var residualRate = 1.0

        let addPayment = fun (roleID: String, address: Address, rate: UFix64) {
            assert(rate >= 0.0 && rate < 1.0, message: "Rate must be in range [0..1)")
            let amount = price * rate

            payments.append(Payment(role: roleID, receiver: address, amount: amount, rate: rate))

            residualRate = residualRate - rate
            assert(residualRate >= 0.0 && residualRate <= 1.0, message: "Residual rate must be in range [0..1)")
        }

        for role in profile.roles.values {
            addPayment(role.id, role.address, role.commissionRate(initialSale: initialSale))
        }

        for role in extraRoles {
            addPayment(role.id, role.address, role.commissionRate(initialSale: initialSale))
        }

        if residualRate > 0.0 {
            addPayment(sellerRole, seller, residualRate)
        }

        return payments
    }
}