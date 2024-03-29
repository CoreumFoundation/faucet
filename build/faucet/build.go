package faucet

import (
	"context"
	"path/filepath"

	"github.com/CoreumFoundation/coreum-tools/pkg/build"
	"github.com/CoreumFoundation/crust/build/golang"
	"github.com/CoreumFoundation/crust/build/tools"
)

const (
	repoPath         = "."
	binaryName       = "faucet"
	binaryPath       = "bin/" + binaryName
	testBinaryPath   = "bin/.cache/integration-tests/faucet"
	goCoverFlag      = "-cover"
	binaryOutputFlag = "-o"
	tagsFlag         = "-tags"
)

// Build builds faucet in docker.
func Build(ctx context.Context, deps build.DepsFunc) error {
	return buildFaucet(ctx, deps, tools.TargetPlatformLinuxLocalArchInDocker, []string{goCoverFlag})
}

func buildFaucet(
	ctx context.Context,
	deps build.DepsFunc,
	targetPlatform tools.TargetPlatform,
	extraFlags []string,
) error {
	binOutputPath := filepath.Join("bin", ".cache", binaryName, targetPlatform.String(), "bin", binaryName)

	return golang.Build(ctx, deps, golang.BinaryBuildConfig{
		TargetPlatform: targetPlatform,
		PackagePath:    repoPath,
		Flags:          append(extraFlags, binaryOutputFlag+"="+binOutputPath),
	})
}

// BuildIntegrationTests builds faucet integration tests.
func BuildIntegrationTests(ctx context.Context, deps build.DepsFunc) error {
	deps(golang.EnsureGo)

	return golang.BuildTests(ctx, golang.TestBuildConfig{
		PackagePath: filepath.Join(repoPath, "integration-tests"),
		Flags: []string{
			binaryOutputFlag + "=" + testBinaryPath,
			tagsFlag + "=" + "integrationtests",
		},
	})
}

// Tidy runs `go mod tidy` for faucet repo.
func Tidy(ctx context.Context, deps build.DepsFunc) error {
	return golang.Tidy(ctx, repoPath, deps)
}

// Lint lints faucet repo.
func Lint(ctx context.Context, deps build.DepsFunc) error {
	return golang.Lint(ctx, repoPath, deps)
}

// Test run unit tests in faucet repo.
func Test(ctx context.Context, deps build.DepsFunc) error {
	return golang.Test(ctx, repoPath, deps)
}
