package iinft

import (
	"errors"
	"os"
	"path"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/emulator"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/gwtf"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
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
	log.Info().Str("filepath", source).Msg("Loading file")
	return f.fsLoader.ReadFile(source)
}

func (f *fileLoader) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return errors.New("file writing not allowed for FlowKit")
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
	var service *services.Services
	if inMemory {
		// YAY, we can run it inline in memory!
		acc, _ := state.EmulatorServiceAccount()
		var gw *emulator.Gateway
		if enableTxFees {
			gw = emulator.NewGatewayWithOpts(acc, emulator.WithTransactionFees())
		} else {
			gw = emulator.NewGatewayWithOpts(acc)
		}
		service = services.NewServices(gw, state, logger)
	} else {
		network, err := state.Networks().ByName(network)
		if err != nil {
			return nil, err
		}
		host := network.Host
		gw, err := gateway.NewGrpcGateway(host)
		if err != nil {
			return nil, err
		}
		service = services.NewServices(gw, state, logger)
	}
	return &gwtf.GoWithTheFlow{
		State:                        state,
		Services:                     service,
		Network:                      network,
		Logger:                       logger,
		PrependNetworkToAccountNames: true,
	}, nil
}
