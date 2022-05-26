import FungibleToken from "./standard/FungibleToken.cdc"
import NonFungibleToken from "./standard/NonFungibleToken.cdc"
import MetadataViews from "./standard/MetadataViews.cdc"

pub contract Evergreen {
    pub struct Role {
        pub let id: String
        pub let description: String
        pub let initialSaleCommission: UFix64
        pub let secondaryMarketCommission: UFix64
        pub let address: Address
        pub let receiverPath: PublicPath?

        init(
            id: String,
            description: String,
            initialSaleCommission: UFix64,
            secondaryMarketCommission: UFix64,
            address: Address,
            receiverPath: PublicPath?
        ) {
            self.id = id
            self.description = description
            self.initialSaleCommission = initialSaleCommission
            self.secondaryMarketCommission = secondaryMarketCommission
            self.address = address
            self.receiverPath = receiverPath
        }

        pub fun commissionRate(initialSale: Bool): UFix64 {
            return (initialSale ? self.initialSaleCommission : self.secondaryMarketCommission)
        }
    }

    pub struct Profile {
        pub let id: UInt32
        pub let description: String  // consider using URI instead
        pub let roles: [Role]

        init(
            id: UInt32,
            description: String,
            roles: [Role]
        ) {
            self.id = id
            self.description = description
            self.roles = roles
        }

        pub fun getRole(id: String): Role? {
            for role in self.roles {
                if (role.id == id) {
                    return role
                }
            }
            return nil
        }

        pub fun buildRoyalties(defaultReceiverPath: PublicPath?): [MetadataViews.Royalty] {
            let royalties: [MetadataViews.Royalty] = []
            for role in self.roles {

                var path = role.receiverPath
                if path == nil {
                    path = defaultReceiverPath
                }

                if path != nil {
                    let receiverCap = getAccount(role.address).getCapability<&{FungibleToken.Receiver}>(path!)
                    if receiverCap.check() {
                        royalties.append(MetadataViews.Royalty(
                            receiver: receiverCap,
                            cut: role.secondaryMarketCommission,
                            description: role.description
                        ))
                    }
                }
            }

            return royalties
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