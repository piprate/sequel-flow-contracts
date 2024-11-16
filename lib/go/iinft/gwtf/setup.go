package gwtf

import (
	"context"
	"fmt"

	"log"

	"github.com/onflow/flow-go-sdk/access"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/gateway"
	"github.com/onflow/flowkit/v2/output"
	"github.com/spf13/afero"
)

// GoWithTheFlow Entire configuration to work with Go With the Flow
type GoWithTheFlow struct {
	State                        *flowkit.State
	Client                       access.Client
	Services                     flowkit.Services
	Logger                       output.Logger
	PrependNetworkToAccountNames bool
}

// NewGoWithTheFlowInMemoryEmulator this method is used to create an in memory emulator, deploy all contracts for the emulator and create all accounts
func NewGoWithTheFlowInMemoryEmulator() *GoWithTheFlow {
	ctx := context.Background()
	return NewGoWithTheFlow(config.DefaultPaths(), "emulator", true, output.InfoLog).InitializeContracts(ctx).CreateAccounts(ctx, "emulator-account")
}

// NewTestingEmulator create new emulator that ignore all log messages
func NewTestingEmulator() *GoWithTheFlow {
	ctx := context.Background()
	return NewGoWithTheFlow(config.DefaultPaths(), "emulator", true, output.NoneLog).InitializeContracts(ctx).CreateAccounts(ctx, "emulator-account")
}

// NewGoWithTheFlowForNetwork creates a new gwtf client for the provided network
func NewGoWithTheFlowForNetwork(network string) *GoWithTheFlow {
	return NewGoWithTheFlow(config.DefaultPaths(), network, false, output.InfoLog)
}

// NewGoWithTheFlowEmulator create a new client
func NewGoWithTheFlowEmulator() *GoWithTheFlow {
	return NewGoWithTheFlow(config.DefaultPaths(), "emulator", false, output.InfoLog)
}

// NewGoWithTheFlowDevNet creates a new gwtf client for devnet/testnet
func NewGoWithTheFlowDevNet() *GoWithTheFlow {
	return NewGoWithTheFlow(config.DefaultPaths(), "testnet", false, output.InfoLog)
}

// NewGoWithTheFlowMainNet creates a new gwft client for mainnet
func NewGoWithTheFlowMainNet() *GoWithTheFlow {
	return NewGoWithTheFlow(config.DefaultPaths(), "mainnet", false, output.InfoLog)
}

// NewGoWithTheFlow with custom file panic on error
func NewGoWithTheFlow(filenames []string, network string, inMemory bool, loglevel int) *GoWithTheFlow {
	gwtf, err := NewGoWithTheFlowError(filenames, network, inMemory, loglevel)
	if err != nil {
		log.Fatalf("error %+v", err)
	}
	return gwtf
}

// DoNotPrependNetworkToAccountNames disable the default behavior of prefixing account names with network-
func (f *GoWithTheFlow) DoNotPrependNetworkToAccountNames() *GoWithTheFlow {
	f.PrependNetworkToAccountNames = false
	return f
}

// Account fetch an account from flow.json, prefixing the name with network- as default (can be turned off)
func (f *GoWithTheFlow) Account(key string) *accounts.Account {
	if f.PrependNetworkToAccountNames {
		key = fmt.Sprintf("%s-%s", f.Services.Network().Name, key)
	}

	account, err := f.State.Accounts().ByName(key)
	if err != nil {
		log.Fatal(err)
	}

	return account
}

// NewGoWithTheFlowError creates a new local go with the flow client
func NewGoWithTheFlowError(paths []string, network string, inMemory bool, logLevel int) (*GoWithTheFlow, error) {

	loader := &afero.Afero{Fs: afero.NewOsFs()}
	state, err := flowkit.Load(paths, loader)
	if err != nil {
		return nil, err
	}

	logger := output.NewStdoutLogger(logLevel)
	var service flowkit.Services
	if inMemory {
		// YAY, we can run it inline in memory!
		acc, _ := state.EmulatorServiceAccount()
		pk, _ := acc.Key.PrivateKey()
		gw := gateway.NewEmulatorGateway(&gateway.EmulatorKey{
			PublicKey: (*pk).PublicKey(),
			SigAlgo:   acc.Key.SigAlgo(),
			HashAlgo:  acc.Key.HashAlgo(),
		})
		service = flowkit.NewFlowkit(state, config.EmulatorNetwork, gw, logger)
	} else {
		networkDef, err := state.Networks().ByName(network)
		if err != nil {
			return nil, err
		}
		gw, err := gateway.NewGrpcGateway(*networkDef)
		if err != nil {
			return nil, err
		}
		service = flowkit.NewFlowkit(state, *networkDef, gw, logger)
	}
	return &GoWithTheFlow{
		State:                        state,
		Services:                     service,
		Logger:                       logger,
		PrependNetworkToAccountNames: true,
	}, nil
}
