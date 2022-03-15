import NonFungibleToken from "./standard/NonFungibleToken.cdc"
import Evergreen from "./Evergreen.cdc"

pub contract DigitalArt: NonFungibleToken {

    // Events
    //
    pub event ContractInitialized()
    pub event Withdraw(id: UInt64, from: Address?)
    pub event Deposit(id: UInt64, to: Address?)
    pub event Minted(id: UInt64, asset: String, edition: UInt64, modID: UInt64)

    // Named Paths
    //
    pub let CollectionStoragePath: StoragePath
    pub let CollectionPublicPath: PublicPath
    pub let AdminStoragePath: StoragePath
    pub let AdminPrivatePath: PrivatePath

    // totalSupply
    // The total number of DigitalArt NFTs that have been minted
    //
    pub var totalSupply: UInt64

    // Variable size dictionary of Master resources
    access(self) var masters: {String: Master}

    pub struct Master {
        pub var metadata: Metadata?
        pub var evergreenProfile: Evergreen.Profile?
        pub var nextEditionId: UInt64
        pub var closed: Bool

        init(metadata: Metadata, evergreenProfile: Evergreen.Profile)  {
            self.metadata = metadata
            self.evergreenProfile = evergreenProfile
            self.nextEditionId = 1
            self.closed = false
        }

        pub fun newEditionID() : UInt64 {
            let val = self.nextEditionId
            self.nextEditionId = self.nextEditionId + UInt64(1)
            return val
        }

        pub fun availableEditions() : UInt64 {
            if !self.closed && self.metadata!.maxEdition >= self.nextEditionId {
                return self.metadata!.maxEdition - self.nextEditionId + UInt64(1)
            } else {
                return 0
            }
        }

        // We close masters after all editions are minted instead of deleting master records
        // This process ensures nobody can ever mint tokens with the same asset ID.
        pub fun close() {
            self.metadata = nil
            self.evergreenProfile = nil
            self.nextEditionId = 0
            self.closed = true
        }
    }

    pub struct Metadata {
        // Link to IPFS file
        pub let metadataLink: String
        // Name
        pub let name: String
        // Artist name
        pub let artist: String
        // Description
        pub let description: String
        // Media type: Audio, Video
        pub let type: String
        pub let contentLink: String
        pub let contentPreviewLink: String
        // MIME type (e.g. 'image/jpeg')
        pub let mimetype: String
		pub var edition: UInt64
		pub let maxEdition: UInt64

		pub let asset: String
		pub let record: String
		pub let assetHead: String

        init(
            metadataLink: String,
            name: String,
            artist: String,
            description: String,
            type: String,
            contentLink: String,
            contentPreviewLink: String,
            mimetype: String,
            edition: UInt64,
            maxEdition: UInt64,
            asset: String,
            record: String,
            assetHead: String
    )  {
            self.metadataLink = metadataLink
            self.name = name
            self.artist = artist
            self.description = description
            self.type = type
            self.contentLink = contentLink
            self.contentPreviewLink = contentPreviewLink
            self.mimetype = mimetype
            self.edition = edition
            self.maxEdition = maxEdition
            self.asset = asset
            self.record = record
            self.assetHead = assetHead
        }

        pub fun setEdition(edition: UInt64) {
            self.edition = edition
        }
    }

    // NFT
    // DigitalArt as an NFT
    //
    pub resource NFT: NonFungibleToken.INFT, Evergreen.Token {
        // The token's ID
        pub let id: UInt64

        pub let metadata: Metadata
        pub let evergreenProfile: Evergreen.Profile

        // initializer
        //
        init(initID: UInt64, metadata: Metadata, evergreenProfile: Evergreen.Profile) {
            self.id = initID
            self.metadata = metadata
            self.evergreenProfile = evergreenProfile
        }

        pub fun getAssetID(): String {
            return self.metadata.asset
        }

        pub fun getEvergreenProfile(): Evergreen.Profile {
            return self.evergreenProfile
        }
    }

    // This is the interface that users can cast their DigitalArt Collection as
    // to allow others to deposit DigitalArt into their Collection. It also allows for reading
    // the details of DigitalArt in the Collection.
    pub resource interface CollectionPublic {
        pub fun deposit(token: @NonFungibleToken.NFT)
        pub fun getIDs(): [UInt64]
        pub fun borrowNFT(id: UInt64): &NonFungibleToken.NFT
        pub fun borrowDigitalArt(id: UInt64): &DigitalArt.NFT? {
            // If the result isn't nil, the id of the returned reference
            // should be the same as the argument to the function
            post {
                (result == nil) || (result?.id == id):
                    "Cannot borrow DigitalArt reference: The ID of the returned reference is incorrect"
            }
        }
    }

    // Collection
    // A collection of DigitalArt NFTs owned by an account
    //
    pub resource Collection: CollectionPublic, Evergreen.CollectionPublic, NonFungibleToken.Provider, NonFungibleToken.Receiver, NonFungibleToken.CollectionPublic {
        // dictionary of NFT conforming tokens
        // NFT is a resource type with an `UInt64` ID field
        //
        pub var ownedNFTs: @{UInt64: NonFungibleToken.NFT}

        // withdraw
        // Removes an NFT from the collection and moves it to the caller
        //
        pub fun withdraw(withdrawID: UInt64): @NonFungibleToken.NFT {
            let token <- self.ownedNFTs.remove(key: withdrawID) ?? panic("missing NFT")

            emit Withdraw(id: token.id, from: self.owner?.address)

            return <-token
        }

        // deposit
        // Takes a NFT and adds it to the collections dictionary
        // and adds the ID to the id array
        //
        pub fun deposit(token: @NonFungibleToken.NFT) {
            let token <- token as! @DigitalArt.NFT

            let id: UInt64 = token.id

            // add the new token to the dictionary which removes the old one
            let oldToken <- self.ownedNFTs[id] <- token

            emit Deposit(id: id, to: self.owner?.address)

            destroy oldToken
        }

        // getIDs
        // Returns an array of the IDs that are in the collection
        //
        pub fun getIDs(): [UInt64] {
            return self.ownedNFTs.keys
        }

        // borrowNFT
        // Gets a reference to an NFT in the collection
        // so that the caller can read its metadata and call its methods
        //
        pub fun borrowNFT(id: UInt64): &NonFungibleToken.NFT {
            return &self.ownedNFTs[id] as &NonFungibleToken.NFT
        }

        // borrowDigitalArt
        // Gets a reference to an NFT in the collection as a DigitalArt,
        // exposing all of its fields (including the typeID).
        // This is safe as there are no functions that can be called on the DigitalArt.
        //
        pub fun borrowDigitalArt(id: UInt64): &DigitalArt.NFT? {
            if self.ownedNFTs[id] != nil {
                let ref = &self.ownedNFTs[id] as auth &NonFungibleToken.NFT
                return ref as! &DigitalArt.NFT
            } else {
                return nil
            }
        }

        pub fun borrowEvergreenToken(id: UInt64): &AnyResource{Evergreen.Token}? {
            return self.borrowDigitalArt(id: id)
        }

        // destructor
        destroy() {
            destroy self.ownedNFTs
        }

        // initializer
        //
        init () {
            self.ownedNFTs <- {}
        }
    }

    // createEmptyCollection
    // public function that anyone can call to create a new empty collection
    //
    pub fun createEmptyCollection(): @NonFungibleToken.Collection {
        return <- create Collection()
    }

    pub fun getMetadata(address:Address, tokenId:UInt64) : Metadata? {
        let acct = getAccount(address)
        let collectionRef = acct.getCapability(self.CollectionPublicPath)!.borrow<&{DigitalArt.CollectionPublic}>()
			?? panic("Could not borrow capability from public collection")

        return collectionRef.borrowDigitalArt(id: tokenId)!.metadata
    }

    // Admin
    // Resource that an admin or something similar would own to be
    // able to mint new NFTs
    //
    pub resource Admin {

        // sealMaster saves and freezes the master copy that then can be used
        // to mint NFT editions.
        pub fun sealMaster(metadata: Metadata, evergreenProfile: Evergreen.Profile) {
            pre {
               metadata.asset != "" : "Empty asset ID"
               metadata.edition == UInt64(0) : "Edition should be zero"
               metadata.maxEdition >= UInt64(1) : "MaxEdition should be positive"
               !DigitalArt.masters.containsKey(metadata.asset) : "master already sealed"
            }
            DigitalArt.masters[metadata.asset] = Master(
                metadata: metadata,
                evergreenProfile: evergreenProfile
            )
        }

        pub fun isSealed(masterId: String) : Bool {
            return DigitalArt.masters.containsKey(masterId)
        }

        pub fun availableEditions(masterId: String) : UInt64 {
            pre {
               DigitalArt.masters.containsKey(masterId) : "master not found"
            }

            let master = &DigitalArt.masters[masterId] as &Master

            return master.availableEditions()
        }

        pub fun evergreenProfile(masterId: String) : Evergreen.Profile {
            pre {
               DigitalArt.masters.containsKey(masterId) : "master not found"
            }

            let master = &DigitalArt.masters[masterId] as &Master

            return master.evergreenProfile!
        }

        pub fun mintEditionNFT(masterId: String, modID: UInt64) : @DigitalArt.NFT {
            pre {
               DigitalArt.masters.containsKey(masterId) : "master not found"
            }

            let master = &DigitalArt.masters[masterId] as &Master

            assert(master.availableEditions() > 0, message: "no more tokens to mint")

            let metadata = master.metadata!
            let edition = master.newEditionID()
            metadata.setEdition(edition: edition)

            // create a new NFT
            var newNFT <- create NFT(
                initID: DigitalArt.totalSupply,
                metadata: metadata,
                evergreenProfile: master.evergreenProfile!
            )

            emit Minted(id: DigitalArt.totalSupply, asset: metadata.asset, edition: edition, modID: modID)

            DigitalArt.totalSupply = DigitalArt.totalSupply + UInt64(1)

            if master.availableEditions() == 0 {
                master.close()
            }

            return <- newNFT
        }
    }

    // initializer
    //
    init() {
        // Set our named paths
        self.CollectionStoragePath = /storage/digitalArtCollection
        self.CollectionPublicPath = /public/digitalArtCollection
        self.AdminStoragePath = /storage/digitalArtAdmin
        self.AdminPrivatePath = /private/digitalArtAdmin

        // Initialize the total supply
        self.totalSupply = 0

        self.masters = {}

        // Create a Admin resource and save it to storage
        let admin <- create Admin()
        self.account.save(<-admin, to: self.AdminStoragePath)

        emit ContractInitialized()
    }
}