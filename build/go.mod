module github.com/CoreumFoundation/faucet/build

go 1.21

// Pin the x/exp dependency version because consmos-sdk breaking change is not compatible
// with cosmos-sdk v0.47.
// Details: https://github.com/cosmos/cosmos-sdk/issues/18415
replace golang.org/x/exp => golang.org/x/exp v0.0.0-20230711153332-06a737ee72cb

require (
	github.com/CoreumFoundation/coreum-tools v0.4.1-0.20240321120602-0a9c50facc68
	github.com/CoreumFoundation/crust/build v0.0.0-20240426160154-ef0a4b5d93ea
	github.com/pkg/errors v0.9.1
)

require (
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/mod v0.16.0 // indirect
)

require (
	github.com/samber/lo v1.39.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	// Make sure to not bump x/exp dependency without cosmos-sdk updated because their breaking change is not compatible
	// with cosmos-sdk v0.47.
	// Details: https://github.com/cosmos/cosmos-sdk/issues/18415
	golang.org/x/exp v0.0.0-20230713183714-613f0c0eb8a1 // indirect
)
