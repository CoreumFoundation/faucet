package main

import (
	coreumTools "github.com/CoreumFoundation/coreum/build/tools"
	"github.com/CoreumFoundation/crust/build"
	"github.com/CoreumFoundation/crust/build/tools"
	selfBuild "github.com/CoreumFoundation/faucet/build"
)

func init() {
	tools.AddTools(coreumTools.Tools...)
}

func main() {
	build.Main(selfBuild.Commands)
}
