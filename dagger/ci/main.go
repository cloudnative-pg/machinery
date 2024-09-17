// A generated module for Ci functions
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

	"dagger/ci/internal/dagger"

	"github.com/sourcegraph/conc/pool"
)

// renovate: datasource=docker depName=golang versioning=semver
const goVersion = "1.23.1"

type Ci struct{}

func (m *Ci) Ci(ctx context.Context, source *dagger.Directory) error {
	_, err := m.ControllerGen(ctx, source).Sync(ctx)
	if err != nil {
		return err
	}

	ctxPool := pool.New().WithContext(ctx)
	ctxPool.Go(func(ctx context.Context) error {
		return m.Test(ctx, source)
	})
	ctxPool.Go(func(ctx context.Context) error {
		return m.CommitLint(ctx, source)
	})
	ctxPool.Go(func(ctx context.Context) error {
		return m.Uncommitted(ctx, source)
	})
	ctxPool.Go(func(ctx context.Context) error {
		return m.SpellCheck(ctx, source)
	})
	ctxPool.Go(func(ctx context.Context) error {
		return m.Lint(ctx, source)
	})

	return ctxPool.Wait()
}

func (m *Ci) Test(ctx context.Context, source *dagger.Directory) error {
	_, err := dag.Gotest().Test(
		ctx,
		dag.Container().From(fmt.Sprintf("golang:%s-alpine", goVersion)),
		source,
	)
	return err
}

func (m *Ci) CommitLint(ctx context.Context, source *dagger.Directory) error {
	ctr := dag.Commitlint().Lint(
		source,
		dagger.CommitlintLintOpts{
			Args: []string{
				"--from=origin/main",
			},
		},
	)
	_, err := ctr.Sync(ctx)
	return err
}

func (m *Ci) Uncommitted(ctx context.Context, source *dagger.Directory) error {
	ctr := dag.Uncommitted().CheckUncommitted(source)
	_, err := ctr.Sync(ctx)
	return err
}

func (m *Ci) ControllerGen(ctx context.Context, source *dagger.Directory) *dagger.Container {
	return dag.ControllerGen().ControllerGen(
		source,
		dagger.ControllerGenControllerGenOpts{
			Args: []string{
				"object:headerFile=hack/boilerplate.go.txt",
				"paths=./pkg/api/...",
			},
		})
}

func (m *Ci) SpellCheck(ctx context.Context, source *dagger.Directory) error {
	_, err := dag.Spellcheck().Spellcheck(source).Sync(ctx)
	return err
}

func (m *Ci) Lint(ctx context.Context, source *dagger.Directory) error {
	_, err := dag.GolangciLint().Run(
		source,
		dagger.GolangciLintRunOpts{
			Config: source.File(".golangci.yml"),
		},
	).Sync(ctx)
	return err
}
