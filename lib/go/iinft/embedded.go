package iinft

import (
	"errors"
	"os"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/internal/assets"
	"github.com/piprate/splash"
)

//go:generate go run github.com/kevinburke/go-bindata/go-bindata -prefix ../../.. -o internal/assets/assets.go -pkg assets -nometadata -nomemcopy ../../../contracts/... ../../../flow.json

type (
	embeddedFileLoader struct {
	}
)

var _ flowkit.ReaderWriter = (*embeddedFileLoader)(nil)

func (f *embeddedFileLoader) ReadFile(source string) ([]byte, error) {
	return assets.Asset(source)
}

func (f *embeddedFileLoader) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return errors.New("operation WriteFile not allowed in embeddedFileLoader")
}

func (f *embeddedFileLoader) MkdirAll(path string, perm os.FileMode) error {
	return errors.New("operation MkdirAll not allowed in embeddedFileLoader")
}

func (f *embeddedFileLoader) Stat(path string) (os.FileInfo, error) {
	return nil, errors.New("operation Stat not allowed in embeddedFileLoader")
}

// NewNetworkConnectorEmbedded creates a new Splash Connector that uses embedded Flow configuration.
func NewNetworkConnectorEmbedded(network string) (*splash.Connector, error) {
	return splash.NewNetworkConnector([]string{config.DefaultPath}, &embeddedFileLoader{}, network, splash.NewZeroLogger())
}

// NewInMemoryConnectorEmbedded creates a new Splash Connector for in-memory emulator that uses embedded Flow configuration.
func NewInMemoryConnectorEmbedded(enableTxFees bool) (*splash.Connector, error) {
	return splash.NewInMemoryConnector([]string{config.DefaultPath}, &embeddedFileLoader{}, enableTxFees, splash.NewZeroLogger())
}
