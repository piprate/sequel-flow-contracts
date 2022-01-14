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
            return (initialSale ? self.initialSaleCommission : self.secondaryMarketCommission)
        }
    }

    pub struct Profile {
        pub let id: UInt32
        pub let roles: { String: Role }

        init(
            id: UInt32,
            roles: { String: Role }
        ) {
            self.id = id
            self.roles = roles
        }
    }

    pub resource interface Token {
        pub fun getAssetID(): String
        pub fun getEvergreenProfile(): Profile
    }

    // An interface for reading the details of an evengreen token in the Collection.
    pub resource interface CollectionPublic {
        pub fun borrowEvergreenToken(id: UInt64): &AnyResource{Token}?
    }
}