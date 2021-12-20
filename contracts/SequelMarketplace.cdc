import FungibleToken from "./standard/FungibleToken.cdc"
import NFTStorefront from "./standard/NFTStorefront.cdc"
import NonFungibleToken from "./standard/NonFungibleToken.cdc"
import Evergreen from "./Evergreen.cdc"
import DigitalArt from "./DigitalArt.cdc"

// SequelMarketplace provides convenience functions to create listings for Sequel NFTs in NFTStorefront.
//
pub contract SequelMarketplace {
    // addListing
    pub fun addListing(
        storefront: &NFTStorefront.Storefront,
        nftProviderCapability: Capability<&{NonFungibleToken.Provider, NonFungibleToken.CollectionPublic, DigitalArt.CollectionPublic}>,
        nftType: Type,
        nftID: UInt64,
        salePaymentVaultPath: PublicPath,
        salePaymentVaultType: Type,
        price: UFix64,
        initialSale: Bool,
        extraRoles: [Evergreen.Role]
    ): UInt64 {
        let token = nftProviderCapability.borrow()!.borrowDigitalArt(id: nftID)!
        let seller = storefront.owner!.address

        let saleCuts = self.buildSaleCuts(
            token: token,
            seller: seller,
            salePaymentVaultPath: salePaymentVaultPath,
            price: price,
            initialSale: initialSale,
            extraRoles: extraRoles)

        let listingID = storefront.createListing(
            nftProviderCapability: nftProviderCapability,
            nftType: nftType,
            nftID: nftID,
            salePaymentVaultType: salePaymentVaultType,
            saleCuts: saleCuts
        )

        return listingID
    }

    pub fun buildSaleCuts(
        token: &AnyResource{Evergreen.Asset},
        seller: Address,
        salePaymentVaultPath: PublicPath,
        price: UFix64,
        initialSale: Bool,
        extraRoles: [Evergreen.Role]
    ): [NFTStorefront.SaleCut] {

        let saleCuts: [NFTStorefront.SaleCut] = []
        var residualRate = 1.0

        let addSaleCut = fun (roleID: String, address: Address, rate: UFix64) {
            assert(rate >= 0.0 && rate < 1.0, message: "Rate must be in range [0..1)")
            let amount = price * rate
            let receiver = getAccount(address).getCapability<&{FungibleToken.Receiver}>(salePaymentVaultPath)
            assert(receiver.borrow() != nil, message: "Missing or mis-typed fungible token receiver")

            saleCuts.append(NFTStorefront.SaleCut(receiver: receiver, amount: amount))

            residualRate = residualRate - rate
            assert(residualRate >= 0.0 && residualRate <= 1.0, message: "Residual rate must be in range [0..1)")
        }

        for role in token.getEvergreenProfile().roles.values {
            addSaleCut(role.id, role.address, role.commissionRate(initialSale: initialSale))
        }

        for role in extraRoles {
            addSaleCut(role.id, role.address, role.commissionRate(initialSale: initialSale))
        }

        addSaleCut("Owner", seller, residualRate)

        return saleCuts
    }
}