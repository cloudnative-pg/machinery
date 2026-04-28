/*
Copyright © contributors to CloudNativePG, established as
CloudNativePG a Series of LF Projects, LLC.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

SPDX-License-Identifier: Apache-2.0
*/

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
	"fmt"

	"dagger/gotest/internal/dagger"
)

type Gotest struct {
	Ctr *dagger.Container
}

func New(
	// Go version
	//
	// +optional
	// +default="latest"
	version string,
	// Container to run the tests
	// +optional
	ctr *dagger.Container,
) *Gotest {
	if ctr != nil {
		return &Gotest{Ctr: ctr}
	}

	user := "noroot"
	modCachePath := fmt.Sprintf("/home/%s/go/pkg/mod", user)
	goCachePath := fmt.Sprintf("/home/%s/.cache/go-build", user)
	ctr = dag.Container().From("golang:"+version).
		WithExec([]string{"useradd", "-m", user}).
		WithUser(user).
		WithEnvVariable("CGO_ENABLED", "0").
		WithEnvVariable("GOMODCACHE", modCachePath).
		WithEnvVariable("GOCACHE", goCachePath).
		WithMountedCache(modCachePath, dag.CacheVolume("go-mod"),
			dagger.ContainerWithMountedCacheOpts{Owner: user}).
		WithMountedCache(goCachePath, dag.CacheVolume("go-build"),
			dagger.ContainerWithMountedCacheOpts{Owner: user})

	return &Gotest{Ctr: ctr}
}

func (m *Gotest) Exec(ctx context.Context, src *dagger.Directory, args ...string) (string, error) {
	return m.Ctr.WithDirectory("/src", src).WithWorkdir("/src").WithExec(args).Stdout(ctx)
}
