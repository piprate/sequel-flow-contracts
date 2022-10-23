module github.com/piprate/sequel-flow-contracts/lib/go/iinft

go 1.16

require (
	github.com/araddon/dateparse v0.0.0-20210429162001-6b43995a97de
	github.com/ethereum/go-ethereum v1.10.21 // indirect
	github.com/klauspost/cpuid v1.2.1 // indirect
	github.com/onflow/cadence v0.28.0
	github.com/onflow/cadence-tools/test v0.2.1-0.20221012182900-f46efb551c55 // indirect
	github.com/onflow/flow-cli/pkg/flowkit v0.0.0-20221012181819-8d43a4be0028
	github.com/onflow/flow-emulator v0.38.1
	github.com/onflow/flow-go v0.28.1-0.20221021192700-74146556857a
	github.com/onflow/flow-go-sdk v0.29.0
	github.com/rs/zerolog v1.26.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/afero v1.9.0
	github.com/stretchr/testify v1.8.0
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/term v0.0.0-20220526004731-065cf7ba2467 // indirect
	google.golang.org/grpc v1.46.2
)

replace github.com/onflow/flow-go-sdk v0.29.0 => github.com/piprate/flow-go-sdk v0.0.0-20221023005443-344f5ca20cea
