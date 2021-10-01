package iinft

import (
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
)

func LoadFlowKitAccount(addrStr, keyStr string) (flowkit.Account, error) {
	acct := flowkit.Account{}

	key, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, keyStr)
	if err != nil {
		return acct, err
	}

	acct.SetName("Sequel")
	acct.SetAddress(flow.HexToAddress(addrStr))
	acct.SetKey(flowkit.NewHexAccountKeyFromPrivateKey(0, crypto.SHA3_256, key))

	return acct, nil
}
