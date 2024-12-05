{{ define "marketplace_withdraw" }}
import NFTStorefront from {{.NFTStorefront}}
import SequelMarketplace from {{.SequelMarketplace}}

transaction(listingID: UInt64) {
    let storefront: auth(NFTStorefront.RemoveListing) &NFTStorefront.Storefront
    let storefrontAddress: Address

    prepare(acct: auth(BorrowValue) &Account) {
        self.storefront = acct.storage.borrow<auth(NFTStorefront.RemoveListing) &NFTStorefront.Storefront>(from: NFTStorefront.StorefrontStoragePath)
            ?? panic("Could not borrow Storefront from provided address")
        self.storefrontAddress = acct.address
    }

    execute {
        SequelMarketplace.withdrawToken(
            storefrontAddress: self.storefrontAddress,
            storefront: self.storefront,
            listingID: listingID,
        )
    }
}
{{ end }}
