package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"dagger/carsapi/internal/dagger"
)

type Carsapi struct{}

const goImage = "golang:1.22-alpine"

// Build all packages in an isolated container.
func (m *Carsapi) Build(
	ctx context.Context,
	// +optional
	// +defaultPath="/"
	// +ignore=[".git"]
	source *dagger.Directory,
) error {
	_, err := m.goEnv(source).
		WithExec([]string{"go", "build", "./..."}).
		Sync(ctx)
	return err
}

// Run all tests in an isolated container with cgo disabled.
func (m *Carsapi) Test(
	ctx context.Context,
	// +optional
	// +defaultPath="/"
	// +ignore=[".git"]
	source *dagger.Directory,
) (string, error) {
	return m.goEnv(source).
		WithEnvVariable("CGO_ENABLED", "0").
		WithExec([]string{"go", "test", "./..."}).
		Stdout(ctx)
}

// Run one package test by exact regex in an isolated container.
func (m *Carsapi) TestOne(
	ctx context.Context,
	// +optional
	// +defaultPath="/"
	// +ignore=[".git"]
	source *dagger.Directory,
	pkg string,
	name string,
) (string, error) {
	return m.goEnv(source).
		WithEnvVariable("CGO_ENABLED", "0").
		WithExec([]string{"go", "test", pkg, "-run", name, "-v"}).
		Stdout(ctx)
}

// Format Go sources in an isolated container and return the updated tree.
func (m *Carsapi) Fmt(
	// +optional
	// +defaultPath="/"
	// +ignore=[".git"]
	source *dagger.Directory,
) *dagger.Directory {
	return m.goEnv(source).
		WithExec([]string{
			"sh",
			"-c",
			"gofmt -w cmd/server/main.go internal/models/*.go internal/repository/*.go internal/service/*.go internal/api/*.go tests/*.go",
		}).
		Directory("/src")
}

// Vet all packages in an isolated container.
func (m *Carsapi) Vet(
	ctx context.Context,
	// +optional
	// +defaultPath="/"
	// +ignore=[".git"]
	source *dagger.Directory,
) error {
	_, err := m.goEnv(source).
		WithExec([]string{"go", "vet", "./..."}).
		Sync(ctx)
	return err
}

// Run the HTTP server as a containerized local service.
func (m *Carsapi) Run(
	// +optional
	// +defaultPath="/"
	// +ignore=[".git"]
	source *dagger.Directory,
	// +optional
	// +default=":8080"
	addr string,
	// +optional
	// +default="cars.db"
	dbPath string,
	// +optional
	// +default="db/schema.sql"
	schemaPath string,
) *dagger.Service {
	return m.goEnv(source).
		WithExec([]string{
			"go",
			"run",
			"./cmd/server",
			"--addr",
			addr,
			"--db-path",
			dbPath,
			"--schema-path",
			schemaPath,
		}).
		WithExposedPort(parsePort(addr)).
		AsService()
}

// Validate core CI checks in isolated containers.
// +check
func (m *Carsapi) Check(
	ctx context.Context,
	// +optional
	// +defaultPath="/"
	// +ignore=[".git"]
	source *dagger.Directory,
) error {
	if err := m.Build(ctx, source); err != nil {
		return err
	}

	if err := m.Vet(ctx, source); err != nil {
		return err
	}

	_, err := m.Test(ctx, source)
	return err
}

func (m *Carsapi) goEnv(source *dagger.Directory) *dagger.Container {
	modCache := dag.CacheVolume("go-mod")
	buildCache := dag.CacheVolume("go-build")

	return dag.Container().
		From(goImage).
		WithMountedCache("/go/pkg/mod", modCache).
		WithMountedCache("/root/.cache/go-build", buildCache).
		WithDirectory("/src", source).
		WithWorkdir("/src")
}

func parsePort(addr string) int {
	parts := strings.Split(addr, ":")
	if len(parts) == 0 {
		return 8080
	}

	port, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		panic(fmt.Sprintf("invalid addr %q: %v", addr, err))
	}

	return port
}
