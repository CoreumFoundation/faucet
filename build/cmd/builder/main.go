package main

import (
	"github.com/CoreumFoundation/crust/build"
	selfBuild "github.com/CoreumFoundation/faucet/build"
)

func main() {
	build.Main(selfBuild.Commands)
}
