package iinft

import (
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flowkit/v2/accounts"
)

func LoadFlowKitAccount(addrStr, keyStr string) (accounts.Account, error) {
	acct := accounts.Account{}

	key, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, keyStr)
	if err != nil {
		return acct, err
	}

	acct.Name = "Sequel"
	acct.Address = flow.HexToAddress(addrStr)
	acct.Key = accounts.NewHexKeyFromPrivateKey(0, crypto.SHA3_256, key)

	return acct, nil
}
