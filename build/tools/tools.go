package tools

import (
	"github.com/CoreumFoundation/crust/build/tools"
)

const (
	// LibWASM is the WASM VM library.
	LibWASM tools.Name = "libwasmvm"
)

// Tools list of required binaries and libraries.
var Tools = []tools.Tool{
	// https://github.com/CosmWasm/wasmvm/releases
	// Check compatibility with wasmd before upgrading: https://github.com/CosmWasm/wasmd
	tools.BinaryTool{
		Name:    LibWASM,
		Version: "v2.2.2",
		Sources: tools.Sources{
			tools.TargetPlatformLinuxAMD64InDocker: {
				URL:  "https://github.com/CosmWasm/wasmvm/releases/download/v2.2.2/libwasmvm_muslc.x86_64.a",
				Hash: "sha256:6dbc82935f204d671392e6dbef0783f48433d3647b76d538430e0888daf048a4",
				Binaries: map[string]string{
					"lib/libwasmvm_muslc.x86_64.a": "libwasmvm_muslc.x86_64.a",
				},
			},
			tools.TargetPlatformLinuxARM64InDocker: {
				URL:  "https://github.com/CosmWasm/wasmvm/releases/download/v2.2.2/libwasmvm_muslc.aarch64.a",
				Hash: "sha256:926ae162b0f7fe3eb35c77e403680c51e7fabc4f8778384bd2ed0b0cb26a6ae2",
				Binaries: map[string]string{
					"lib/libwasmvm_muslc.aarch64.a": "libwasmvm_muslc.aarch64.a",
				},
			},
			tools.TargetPlatformDarwinAMD64InDocker: {
				URL:  "https://github.com/CosmWasm/wasmvm/releases/download/v2.2.2/libwasmvmstatic_darwin.a",
				Hash: "sha256:3de037b934e682dec05c5ec4f0378b62b1b2444627c609d8821e00d126cd409b",
				Binaries: map[string]string{
					"lib/libwasmvmstatic_darwin.a": "libwasmvmstatic_darwin.a",
				},
			},
			tools.TargetPlatformDarwinARM64InDocker: {
				URL:  "https://github.com/CosmWasm/wasmvm/releases/download/v2.2.2/libwasmvmstatic_darwin.a",
				Hash: "sha256:3de037b934e682dec05c5ec4f0378b62b1b2444627c609d8821e00d126cd409b",
				Binaries: map[string]string{
					"lib/libwasmvmstatic_darwin.a": "libwasmvmstatic_darwin.a",
				},
			},
		},
	},
}
