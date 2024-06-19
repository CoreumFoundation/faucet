package build

import (
	"github.com/CoreumFoundation/crust/build/crust"
	"github.com/CoreumFoundation/crust/build/golang"
	"github.com/CoreumFoundation/crust/build/types"
	"github.com/CoreumFoundation/faucet/build/faucet"
)

// Commands is a definition of commands available in build system.
var Commands = map[string]types.Command{
	"build/me":          {Fn: crust.BuildBuilder, Description: "Builds the builder"},
	"build/znet":        {Fn: crust.BuildZNet, Description: "Builds znet binary"},
	"build":             {Fn: faucet.Build, Description: "Builds faucet binary"},
	"images":            {Fn: faucet.BuildDockerImage, Description: "Builds faucet docker image"},
	"integration-tests": {Fn: faucet.RunIntegrationTests, Description: "Runs integration tests"},
	"lint":              {Fn: golang.Lint, Description: "Lints code"},
	"release":           {Fn: faucet.Release, Description: "Releases faucet binary"},
	"release/images":    {Fn: faucet.ReleaseImage, Description: "Releases faucet docker images"},
	"test":              {Fn: golang.Test, Description: "Runs unit tests"},
	"tidy":              {Fn: golang.Tidy, Description: "Runs go mod tidy"},
}
