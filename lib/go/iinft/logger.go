package iinft

import (
	"strings"

	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type (
	logger struct {
		log zerolog.Logger
	}
)

var _ output.Logger = (*logger)(nil)

func NewFlowKitLogger() output.Logger {
	zeroLogger := log.Logger.With().Str("module", "flow").Logger()

	return &logger{
		log: zeroLogger,
	}
}

func (l logger) Debug(s string) {
	log.Debug().Msg(stripLineBreaks(s))
}

func (l logger) Info(s string) {
	log.Info().Msg(stripLineBreaks(s))
}

func (l logger) Error(s string) {
	log.Error().Msg(stripLineBreaks(s))
}

func (l logger) StartProgress(s string) {
	log.Debug().Msg(stripLineBreaks(s))
}

func (l logger) StopProgress() {
	// do nothing
}

func stripLineBreaks(s string) string {
	return strings.Trim(s, "\n")
}
