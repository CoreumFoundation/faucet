package build

import (
	"github.com/CoreumFoundation/coreum-tools/pkg/build"
	"github.com/CoreumFoundation/crust/build/crust"
	"github.com/CoreumFoundation/faucet/build/faucet"
)

// Commands is a definition of commands available in build system.
var Commands = map[string]build.Command{
	"build/me":          {Fn: crust.BuildBuilder, Description: "Builds the builder"},
	"build":             {Fn: faucet.Build, Description: "Builds faucet binary"},
	"download":          {Fn: faucet.DownloadDependencies, Description: "Downloads go dependencies"},
	"images":            {Fn: faucet.BuildDockerImage, Description: "Builds faucet docker image"},
	"integration-tests": {Fn: faucet.RunIntegrationTests, Description: "Runs integration tests"},
	"lint":              {Fn: faucet.Lint, Description: "Lints code"},
	"release":           {Fn: faucet.Release, Description: "Releases faucet binary"},
	"release/images":    {Fn: faucet.ReleaseImage, Description: "Releases faucet docker images"},
	"test":              {Fn: faucet.Test, Description: "Runs unit tests"},
	"tidy":              {Fn: faucet.Tidy, Description: "Runs go mod tidy"},
}
