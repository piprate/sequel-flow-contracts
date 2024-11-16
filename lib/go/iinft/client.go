package iinft

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/onflow/flow-emulator/emulator"
	"github.com/onflow/flow-go-sdk/access"
	grpcAccess "github.com/onflow/flow-go-sdk/access/grpc"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/gateway"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/gwtf"
	"github.com/spf13/afero"
	"google.golang.org/grpc"
)

type (
	fileLoader struct {
		baseDir  string
		fsLoader *afero.Afero
	}
)

var _ flowkit.ReaderWriter = (*fileLoader)(nil)

func (f *fileLoader) ReadFile(source string) ([]byte, error) {
	source = path.Join(f.baseDir, source)
	return f.fsLoader.ReadFile(source)
}

func (f *fileLoader) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return errors.New("file writing not allowed in fileLoader")
}

func (f *fileLoader) MkdirAll(path string, perm os.FileMode) error {
	return errors.New("directory creation not allowed in fileLoader")
}

func (f *fileLoader) Stat(path string) (os.FileInfo, error) {
	return nil, errors.New("operation Stat not supported in fileLoader")
}

// NewGoWithTheFlowFS creates a new local go with the flow client
func NewGoWithTheFlowFS(flowBasePath, network string, inMemory, enableTxFees bool) (*gwtf.GoWithTheFlow, error) {
	return NewGoWithTheFlowError(&fileLoader{
		baseDir:  flowBasePath,
		fsLoader: &afero.Afero{Fs: afero.NewOsFs()},
	}, network, inMemory, enableTxFees)
}

// NewGoWithTheFlowEmbedded creates a new test go with the flow client based on embedded setup
func NewGoWithTheFlowEmbedded(network string, inMemory, enableTxFees bool) (*gwtf.GoWithTheFlow, error) {
	return NewGoWithTheFlowError(&embeddedFileLoader{}, network, inMemory, enableTxFees)
}

func NewGoWithTheFlowError(baseLoader flowkit.ReaderWriter, network string, inMemory, enableTxFees bool) (*gwtf.GoWithTheFlow, error) {

	state, err := flowkit.Load([]string{"flow.json"}, baseLoader)
	if err != nil {
		return nil, err
	}

	logger := NewFlowKitLogger()
	var service *flowkit.Flowkit
	var client access.Client

	if inMemory {
		// YAY, we can run it inline in memory!
		acc, _ := state.EmulatorServiceAccount()
		pk, _ := acc.Key.PrivateKey()
		var gw *gateway.EmulatorGateway
		if enableTxFees {
			gw = gateway.NewEmulatorGatewayWithOpts(&gateway.EmulatorKey{
				PublicKey: (*pk).PublicKey(),
				SigAlgo:   acc.Key.SigAlgo(),
				HashAlgo:  acc.Key.HashAlgo(),
			}, gateway.WithEmulatorOptions(emulator.WithTransactionFeesEnabled(true)))
		} else {
			gw = gateway.NewEmulatorGateway(&gateway.EmulatorKey{
				PublicKey: (*pk).PublicKey(),
				SigAlgo:   acc.Key.SigAlgo(),
				HashAlgo:  acc.Key.HashAlgo(),
			})
		}
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

		client, err = grpcAccess.NewClient(
			networkDef.Host,
			grpcAccess.WithGRPCDialOptions(
				grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxGRPCMessageSize)),
			),
		)

		if err != nil || client == nil {
			return nil, fmt.Errorf("failed to connect to host %s", networkDef.Host)
		}
	}
	return &gwtf.GoWithTheFlow{
		State:                        state,
		Services:                     service,
		Client:                       client,
		Logger:                       logger,
		PrependNetworkToAccountNames: true,
	}, nil
}

// NewGrpcClientForNetworkFS creates a new local go with the flow client
func NewGrpcClientForNetworkFS(flowBasePath, network string) (access.Client, error) {
	return NewGrpcClient(&fileLoader{
		baseDir:  flowBasePath,
		fsLoader: &afero.Afero{Fs: afero.NewOsFs()},
	}, network)
}

// NewGrpcClientForNetworkEmbedded creates a new test go with the flow client based on embedded setup
func NewGrpcClientForNetworkEmbedded(network string) (access.Client, error) {
	return NewGrpcClient(&embeddedFileLoader{}, network)
}

// maxGRPCMessageSize 60mb
const maxGRPCMessageSize = 1024 * 1024 * 60

func NewGrpcClient(baseLoader flowkit.ReaderWriter, network string, opts ...grpcAccess.ClientOption) (access.Client, error) {
	state, err := flowkit.Load([]string{"flow.json"}, baseLoader)
	if err != nil {
		return nil, err
	}

	networkDef, err := state.Networks().ByName(network)
	if err != nil {
		return nil, err
	}

	options := append(
		[]grpcAccess.ClientOption{
			grpcAccess.WithGRPCDialOptions(
				grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxGRPCMessageSize)),
			),
		},
		opts...,
	)

	gClient, err := grpcAccess.NewClient(
		networkDef.Host,
		options...,
	)

	if err != nil || gClient == nil {
		return nil, fmt.Errorf("failed to connect to host %s", networkDef.Host)
	}

	return gClient, nil
}
