package main

import (
	"github.com/CoreumFoundation/crust/build"
	"github.com/CoreumFoundation/crust/build/tools"
	selfBuild "github.com/CoreumFoundation/faucet/build"
	selfTools "github.com/CoreumFoundation/faucet/build/tools"
)

func init() {
	tools.AddTools(selfTools.Tools...)
}

func main() {
	build.Main(selfBuild.Commands)
}
