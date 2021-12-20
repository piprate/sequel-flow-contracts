import FungibleToken from "./standard/FungibleToken.cdc"
import NFTStorefront from "./standard/NFTStorefront.cdc"
import NonFungibleToken from "./standard/NonFungibleToken.cdc"
import Evergreen from "./Evergreen.cdc"
import DigitalArt from "./DigitalArt.cdc"

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
        payments: [Payment]
    )

    pub event TokenSold(
        storefrontAddress: Address,
        listingID: UInt64,
        nftType: String,
        nftID: UInt64,
        paymentVaultType: String,
        price: UFix64,
        buyerAddress: Address
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
        nftProviderCapability: Capability<&{NonFungibleToken.Provider, NonFungibleToken.CollectionPublic, DigitalArt.CollectionPublic}>,
        nftType: Type,
        nftID: UInt64,
        paymentVaultPath: PublicPath,
        paymentVaultType: Type,
        price: UFix64,
        initialSale: Bool,
        extraRoles: [Evergreen.Role]
    ): UInt64 {
        let token = nftProviderCapability.borrow()!.borrowDigitalArt(id: nftID)!
        let seller = storefront.owner!.address

        let payments = self.buildPayments(
            token: token,
            seller: seller,
            price: price,
            initialSale: initialSale,
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
            payments: payments
        )

        return listingID
    }

    pub fun buyToken(
        storefrontAddress: Address,
        storefront: &NFTStorefront.Storefront{NFTStorefront.StorefrontPublic},
        listingID: UInt64,
        listing: &NFTStorefront.Listing{NFTStorefront.ListingPublic},
        paymentVault: @FungibleToken.Vault,
        buyerAddress: Address
    ): @NonFungibleToken.NFT {
        let details = listing.getDetails()

        emit TokenSold(
            storefrontAddress: storefrontAddress,
            listingID: listingID,
            nftType: details.nftType.identifier,
            nftID: details.nftID,
            paymentVaultType: details.salePaymentVaultType.identifier,
            price: details.salePrice,
            buyerAddress: buyerAddress
        )

        let item <- listing.purchase(payment: <-paymentVault)
        storefront.cleanup(listingResourceID: listingID)
        return <- item
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
        token: &AnyResource{Evergreen.Asset},
        seller: Address,
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

        for role in token.getEvergreenProfile().roles.values {
            addPayment(role.id, role.address, role.commissionRate(initialSale: initialSale))
        }

        for role in extraRoles {
            addPayment(role.id, role.address, role.commissionRate(initialSale: initialSale))
        }

        addPayment("Owner", seller, residualRate)

        return payments
    }
}