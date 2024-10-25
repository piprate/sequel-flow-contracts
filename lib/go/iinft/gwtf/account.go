package gwtf

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
)

func (f *GoWithTheFlow) CreateAccounts(ctx context.Context, saAccountName string) *GoWithTheFlow {
	gwtf, err := f.CreateAccountsE(ctx, saAccountName)
	if err != nil {
		log.Fatal(err)
	}

	return gwtf
}

// CreateAccountsE ensures that all accounts present in the deployment block for the given network is present
func (f *GoWithTheFlow) CreateAccountsE(ctx context.Context, saAccountName string) (*GoWithTheFlow, error) {
	p := f.State
	signerAccount, err := p.Accounts().ByName(saAccountName)
	if err != nil {
		return nil, err
	}

	accountList := *p.AccountsForNetwork(f.Services.Network())
	accountNames := accountList.Names()
	sort.Strings(accountNames)

	f.Logger.Info(fmt.Sprintf("%v\n", accountNames))

	for _, accountName := range accountNames {
		f.Logger.Debug(fmt.Sprintf("Ensuring account with name '%s' is present", accountName))

		// this error can never happen here, there is a test for it.
		account, _ := p.Accounts().ByName(accountName)

		if _, err := f.Services.GetAccount(ctx, account.Address); err == nil {
			f.Logger.Debug("Account is present")
			continue
		}

		a, _, err := f.Services.CreateAccount(
			ctx,
			signerAccount,
			[]accounts.PublicKey{{
				Public:   account.Key.ToConfig().PrivateKey.PublicKey(),
				Weight:   flow.AccountKeyWeightThreshold,
				SigAlgo:  account.Key.SigAlgo(),
				HashAlgo: account.Key.HashAlgo(),
			}})
		if err != nil {
			return nil, err
		}
		f.Logger.Info("Account created " + a.Address.String())
		if a.Address.String() != account.Address.String() {
			// this condition happens when we create accounts defined in flow.json
			// after some other accounts were created manually.
			// In this case, account addresses may not match the expected values
			f.Logger.Error("Account address mismatch. Expected " + account.Address.String() + ", got " + a.Address.String())
		}
	}
	return f, nil
}

// InitializeContracts installs all contracts in the deployment block for the configured network
func (f *GoWithTheFlow) InitializeContracts(ctx context.Context) *GoWithTheFlow {
	if err := f.InitializeContractsE(ctx); err != nil {
		log.Fatal(err)
	}

	return f
}

// InitializeContractsE installs all contracts in the deployment block for the configured network
// and returns an error if it fails.
func (f *GoWithTheFlow) InitializeContractsE(ctx context.Context) error {
	f.Logger.Info("Deploying contracts")
	if _, err := f.Services.DeployProject(ctx, flowkit.UpdateExistingContract(true)); err != nil {
		return err
	}

	return nil
}
