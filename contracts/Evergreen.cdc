import FungibleToken from "./standard/FungibleToken.cdc"
import NFTStorefront from "./standard/NFTStorefront.cdc"
import NonFungibleToken from "./standard/NonFungibleToken.cdc"

pub contract Evergreen {
    pub struct Role {
        pub let id: String
        pub let initialSaleCommission: UFix64
        pub let secondaryMarketCommission: UFix64
        pub let address: Address

        init(
            id: String,
            initialSaleCommission: UFix64,
            secondaryMarketCommission: UFix64,
            address: Address
        ) {
            self.id = id
            self.initialSaleCommission = initialSaleCommission
            self.secondaryMarketCommission = secondaryMarketCommission
            self.address = address
        }

        pub fun commissionRate(initialSale: Bool): UFix64 {
            return (initialSale ? self.initialSaleCommission : self.secondaryMarketCommission) / 100.0
        }
    }

    pub struct Profile {
        pub let id: UInt32
        pub let roles: { String: Role }
        pub let description: String

        init(
            id: UInt32,
            roles: { String: Role }
            description: String
        ) {
            self.id = id
            self.roles = roles
            self.description = description
        }
    }

    pub resource interface Asset {
        pub fun getEvergreenProfile(): Profile
    }
}