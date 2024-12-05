package iinft

import (
	"github.com/onflow/flowkit/v2/config"
	contracts "github.com/piprate/sequel-flow-contracts"
	"github.com/piprate/splash"
)

// NewNetworkConnectorEmbedded creates a new Splash Connector that uses embedded Flow configuration.
func NewNetworkConnectorEmbedded(network string) (*splash.Connector, error) {
	return splash.NewNetworkConnector([]string{config.DefaultPath}, splash.NewEmbedLoader(&contracts.ResourcesFS), network, splash.NewZeroLogger())
}

// NewInMemoryConnectorEmbedded creates a new Splash Connector for in-memory emulator that uses embedded Flow configuration.
func NewInMemoryConnectorEmbedded(enableTxFees bool) (*splash.Connector, error) {
	return splash.NewInMemoryConnector([]string{config.DefaultPath}, splash.NewEmbedLoader(&contracts.ResourcesFS), enableTxFees, splash.NewZeroLogger())
}
