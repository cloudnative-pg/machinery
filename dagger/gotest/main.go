// A generated module for Gotest functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"

	"dagger/gotest/internal/dagger"
)

type Gotest struct{}

func (m *Gotest) Test(ctx context.Context, ctr *dagger.Container, src *dagger.Directory) (string, error) {
	ctr = ctr.WithExec([]string{"apk", "add", "--no-cache", "git", "curl", "unzip"}).
		WithExec([]string{"adduser", "-D", "-h", "/home/user", "-u", "1000", "user"}).WithUser("1000").
		WithMountedCache("/home/user/go/pkg/mod", dag.CacheVolume("go-mod-123"),
			dagger.ContainerWithMountedCacheOpts{Owner: "user"},
		).
		WithEnvVariable("GOMODCACHE", "/home/user/go/pkg/mod").
		WithMountedCache("/home/user/go/build-cache", dag.CacheVolume("go-build-123"),
			dagger.ContainerWithMountedCacheOpts{Owner: "user"},
		).
		WithEnvVariable("GOCACHE", "/home/user/go/build-cache")

	return dag.Gotoolbox(dagger.GotoolboxOpts{Ctr: ctr}).WithCgoDisabled().RunGoCmd(ctx,
		[]string{"test", "-v", "./..."},
		dagger.GotoolboxRunGoCmdOpts{
			Src: src,
		})
}
