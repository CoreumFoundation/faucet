package main

import (
	"github.com/CoreumFoundation/coreum-tools/pkg/build"
	selfBuild "github.com/CoreumFoundation/faucet/build"
)

func main() {
	build.Main(selfBuild.Commands)
}
