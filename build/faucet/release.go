package faucet

import (
	"context"

	"github.com/pkg/errors"

	"github.com/CoreumFoundation/crust/build/config"
	"github.com/CoreumFoundation/crust/build/docker"
	"github.com/CoreumFoundation/crust/build/git"
	"github.com/CoreumFoundation/crust/build/tools"
	"github.com/CoreumFoundation/crust/build/types"
)

// Release releases faucet binary for amd64 and arm64 to be published inside the release.
func Release(ctx context.Context, deps types.DepsFunc) error {
	clean, _, err := git.StatusClean(ctx)
	if err != nil {
		return err
	}
	if !clean {
		return errors.New("released commit contains uncommitted changes")
	}

	version, err := git.VersionFromTag(ctx)
	if err != nil {
		return err
	}
	if version == "" {
		return errors.New("no version present on released commit")
	}

	if err := buildFaucet(ctx, deps, tools.TargetPlatformLinuxAMD64InDocker, []string{}); err != nil {
		return err
	}
	return buildFaucet(ctx, deps, tools.TargetPlatformLinuxARM64InDocker, []string{})
}

// ReleaseImage releases faucet docker images for amd64 and arm64.
func ReleaseImage(ctx context.Context, deps types.DepsFunc) error {
	deps(Release)

	return buildDockerImage(ctx, imageConfig{
		TargetPlatforms: []tools.TargetPlatform{
			tools.TargetPlatformLinuxAMD64InDocker,
			tools.TargetPlatformLinuxARM64InDocker,
		},
		Action:   docker.ActionPush,
		Username: config.DockerHubUsername,
	})
}
