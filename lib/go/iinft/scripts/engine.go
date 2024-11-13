package scripts

import (
	"bytes"
	"embed"
	"fmt"

	"path"
	"strings"
	"text/template"

	"github.com/onflow/cadence/format"
	"github.com/onflow/flow-go-sdk"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft"
	"github.com/piprate/sequel-flow-contracts/lib/go/iinft/gwtf"
	"github.com/rs/zerolog/log"
)

//go:embed templates
var templateFS embed.FS
var goTemplates *template.Template

func init() {
	var err error
	goTemplates, err = template.New("").Funcs(template.FuncMap{
		// increment function
		"inc": func(i int) int {
			return i + 1
		},
		// decrement function
		"dec": func(i int) int {
			return i - 1
		},
		// turn a string into Cadence safe form
		"safe": func(v string) string {
			return format.String(v)
		},
		"ufix64": iinft.UFix64ToString,
	}).ParseFS(templateFS, "templates/transactions/*.cdc", "templates/scripts/*.cdc", "templates/scripts/**/*.cdc")
	if err != nil {
		panic(err)
	}
}

type (
	Engine struct {
		client                   *gwtf.GoWithTheFlow
		preloadedTemplates       map[string]string
		wellKnownAddresses       map[string]string
		wellKnownAddressesBinary map[string]flow.Address
	}
)

const (
	ParamsKey = "Parameters"
)

var (
	requiredWellKnownContracts = []string{
		"Burner", "FlowToken", "FungibleToken", "FungibleTokenMetadataViews",
		"FungibleTokenSwitchboard", "MetadataViews", "NFTStorefront",
		"NonFungibleToken", "NFTCatalog", "NFTRetrieval",
		"Art", "Content",
		"Evergreen", "DigitalArt", "SequelMarketplace",
	}
)

func NewEngine(client *gwtf.GoWithTheFlow, preload bool) (*Engine, error) {
	eng := &Engine{
		client:                   client,
		preloadedTemplates:       make(map[string]string),
		wellKnownAddresses:       make(map[string]string),
		wellKnownAddressesBinary: make(map[string]flow.Address),
	}

	if err := eng.loadContractAddresses(); err != nil {
		return nil, err
	}

	return eng, nil
}

func (e *Engine) loadContractAddresses() error {
	contracts := e.client.State.Contracts()
	network := e.client.Services.Network()
	networkName := network.Name
	deployedContracts, err := e.client.State.DeploymentContractsByNetwork(network)
	if err != nil {
		return err
	}
	for _, contract := range *contracts {
		for _, alias := range contract.Aliases {
			if alias.Network == networkName {
				e.wellKnownAddressesBinary[contract.Name] = alias.Address
			}
		}
	}
	for _, contract := range deployedContracts {
		e.wellKnownAddressesBinary[strings.Split(path.Base(contract.Location()), ".")[0]] = contract.AccountAddress
	}

	for _, requiredContractName := range requiredWellKnownContracts {
		if _, found := e.wellKnownAddressesBinary[requiredContractName]; !found {
			return fmt.Errorf("address not found for contract %s", requiredContractName)
		}
	}
	log.Debug().Str("addresses", fmt.Sprintf("%v", e.wellKnownAddresses)).Msg("Loaded contract addresses")

	for name, addr := range e.wellKnownAddressesBinary {
		e.wellKnownAddresses[name] = addr.HexWithPrefix()
	}

	return nil
}

func (e *Engine) WellKnownAddresses() map[string]string {
	return e.wellKnownAddresses
}

func (e *Engine) ContractAddress(contractName string) flow.Address {
	return e.wellKnownAddressesBinary[contractName]
}

func (e *Engine) GetStandardScript(scriptID string) string {
	s, found := e.preloadedTemplates[scriptID]
	if !found {
		buf := &bytes.Buffer{}
		if err := goTemplates.ExecuteTemplate(buf, scriptID, e.wellKnownAddresses); err != nil {
			panic(err)
		}

		s = buf.String()
		e.preloadedTemplates[scriptID] = s
	}

	return s
}

func (e *Engine) GetCustomScript(scriptID string, params interface{}) string {
	data := map[string]interface{}{
		ParamsKey: params,
	}
	for k, v := range e.wellKnownAddresses {
		data[k] = v
	}
	buf := &bytes.Buffer{}
	if err := goTemplates.ExecuteTemplate(buf, scriptID, data); err != nil {
		panic(err)
	}

	return buf.String()
}

func (e *Engine) NewTransaction(scriptID string) gwtf.FlowTransactionBuilder {
	return e.client.Transaction(e.GetStandardScript(scriptID))
}

func (e *Engine) NewInlineTransaction(script string) gwtf.FlowTransactionBuilder {
	return e.client.Transaction(script)
}

func (e *Engine) NewScript(scriptID string) gwtf.FlowScriptBuilder {
	return e.client.Script(e.GetStandardScript(scriptID))
}

func (e *Engine) NewInlineScript(script string) gwtf.FlowScriptBuilder {
	return e.client.Script(script)
}
