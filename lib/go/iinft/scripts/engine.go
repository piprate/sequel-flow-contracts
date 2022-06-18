package scripts

import (
	"bytes"
	"embed"
	"fmt"
	"path"
	"strings"
	"text/template"

	"github.com/onflow/cadence/runtime/format"
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
	}).ParseFS(templateFS, "templates/transactions/*.cdc", "templates/scripts/*.cdc", "templates/scripts/**/*.cdc")
	if err != nil {
		panic(err)
	}
}

type (
	Engine struct {
		client             *gwtf.GoWithTheFlow
		preloadedTemplates map[string]string
		wellKnownAddresses map[string]string
	}
)

const (
	ParamsKey = "Parameters"
)

var (
	requiredWellKnownAddresses = []string{
		"FungibleToken", "FlowToken", "NonFungibleToken", "NFTStorefront", "FUSD",
		"Collectible", "Edition", "Art", "Content", "Evergreen",
		"DigitalArt", "SequelMarketplace",
	}
)

func NewEngine(client *gwtf.GoWithTheFlow, preload bool) (*Engine, error) {
	eng := &Engine{
		client:             client,
		preloadedTemplates: make(map[string]string),
		wellKnownAddresses: make(map[string]string),
	}

	if err := eng.loadContractAddresses(); err != nil {
		return nil, err
	}

	return eng, nil
}

func (e *Engine) loadContractAddresses() error {
	contracts := e.client.State.Contracts().ByNetwork(e.client.Network)
	deployedContracts, err := e.client.State.DeploymentContractsByNetwork(e.client.Network)
	if err != nil {
		return err
	}
	for _, contract := range contracts {
		if contract.Alias != "" {
			e.wellKnownAddresses[contract.Name] = contract.Alias
		}
	}
	for _, contract := range deployedContracts {
		e.wellKnownAddresses[strings.Split(path.Base(contract.Source), ".")[0]] = "0x" + contract.AccountAddress.String()
	}

	for _, requiredAddress := range requiredWellKnownAddresses {
		if _, found := e.wellKnownAddresses[requiredAddress]; !found {
			return fmt.Errorf("address not found for contract %s", requiredAddress)
		}
	}
	log.Debug().Str("addresses", fmt.Sprintf("%v", e.wellKnownAddresses)).Msg("Loaded contract addresses")

	return nil
}

func (e *Engine) WellKnownAddresses() map[string]string {
	return e.wellKnownAddresses
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
