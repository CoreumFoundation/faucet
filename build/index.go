package build

import (
	"github.com/CoreumFoundation/coreum-tools/pkg/build"
	"github.com/CoreumFoundation/crust/build/crust"
	"github.com/CoreumFoundation/faucet/build/faucet"
)

// Commands is a definition of commands available in build system.
var Commands = map[string]build.CommandFunc{
	"build/me":                crust.BuildBuilder,
	"build":                   faucet.Build,
	"build/integration-tests": faucet.BuildIntegrationTests,
	"images":                  faucet.BuildDockerImage,
	"lint":                    faucet.Lint,
	"release":                 faucet.Release,
	"release/images":          faucet.ReleaseImage,
	"test":                    faucet.Test,
	"tidy":                    faucet.Tidy,
}
