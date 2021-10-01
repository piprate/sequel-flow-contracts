package test

import (
	"errors"
	"os"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/test/internal/assets"
	"github.com/rs/zerolog/log"
)

//go:generate go run github.com/kevinburke/go-bindata/go-bindata -prefix ../../../.. -o internal/assets/assets.go -pkg assets -nometadata -nomemcopy ../../../../contracts/... ../../../../flow.json

type (
	embeddedFileLoader struct {
	}
)

var _ flowkit.ReaderWriter = (*embeddedFileLoader)(nil)

func (f *embeddedFileLoader) ReadFile(source string) ([]byte, error) {
	log.Info().Str("filepath", source).Msg("Loading embedded file")
	return assets.Asset(source)
}

func (f *embeddedFileLoader) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return errors.New("file writing not allowed for FlowKit")
}
