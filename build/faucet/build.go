package faucet

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/samber/lo"

	"github.com/CoreumFoundation/coreum/build/coreum"
	"github.com/CoreumFoundation/crust/build/golang"
	"github.com/CoreumFoundation/crust/build/tools"
	"github.com/CoreumFoundation/crust/build/types"
	"github.com/CoreumFoundation/crust/infra"
	"github.com/CoreumFoundation/crust/infra/apps"
	"github.com/CoreumFoundation/crust/pkg/znet"
)

const (
	repoPath    = "."
	binaryName  = "faucet"
	binaryPath  = "bin/" + binaryName
	goCoverFlag = "-cover"
)

// Build builds faucet in docker.
func Build(ctx context.Context, deps types.DepsFunc) error {
	return buildFaucet(ctx, deps, tools.TargetPlatformLinuxLocalArchInDocker, []string{goCoverFlag})
}

func buildFaucet(
	ctx context.Context,
	deps types.DepsFunc,
	targetPlatform tools.TargetPlatform,
	extraFlags []string,
) error {
	binOutputPath := filepath.Join("bin", ".cache", binaryName, targetPlatform.String(), "bin", binaryName)

	return golang.Build(ctx, deps, golang.BinaryBuildConfig{
		TargetPlatform: targetPlatform,
		PackagePath:    "cmd",
		BinOutputPath:  binOutputPath,
		Flags:          extraFlags,
	})
}

// RunIntegrationTests runs faucet integration tests.
func RunIntegrationTests(ctx context.Context, deps types.DepsFunc) (retErr error) {
	deps(BuildDockerImage, coreum.BuildCoredLocally, coreum.BuildCoredDockerImage)

	znetConfig := &infra.ConfigFactory{
		EnvName:       "znet",
		TimeoutCommit: 500 * time.Millisecond,
		HomeDir:       filepath.Join(lo.Must(os.UserHomeDir()), ".crust", "znet"),
		RootDir:       ".",
		Profiles:      []string{apps.Profile1Cored, apps.ProfileFaucet},
	}

	if err := znet.Remove(ctx, znetConfig); err != nil {
		return err
	}
	defer func() {
		if err := znet.Remove(ctx, znetConfig); retErr == nil {
			retErr = err
		}
	}()

	if err := znet.Start(ctx, znetConfig); err != nil {
		return err
	}
	return golang.RunTests(ctx, deps, golang.TestConfig{
		PackagePath: filepath.Join(repoPath, "integration-tests"),
		Flags: []string{
			"-tags=integrationtests",
			fmt.Sprintf("-parallel=%d", 2*runtime.NumCPU()),
			"-timeout=1h",
		},
	})
}
