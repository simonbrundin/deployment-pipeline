package main

import (
	"context"
	"fmt"

	"dagger.io/dagger"
)

type CiPipeline struct{}

// TestAndBuild kör unit tester och bygger container image med Buildah
func (m *CiPipeline) TestAndBuild(ctx context.Context, source *dagger.Directory, imageName string) (string, error) {
	if imageName == "" {
		imageName = "myapp"
	}

	// Skapa en container med nödvändiga verktyg
	container := dag.Container().
		From("registry.fedoraproject.org/fedora:latest").
		WithExec([]string{"dnf", "install", "-y", "buildah", "golang", "git"}).
		WithDirectory("/src", source).
		WithWorkdir("/src")

	// Kör Go tester
	testContainer := container.
		WithExec([]string{"go", "mod", "download"}).
		WithExec([]string{"go", "test", "-v", "./..."}).
		WithExec([]string{"go", "build", "-o", "app", "."})

	// Bygg image med Buildah
	buildContainer := testContainer.
		WithExec([]string{"buildah", "bud", "-t", imageName, "."}).
		WithExec([]string{"buildah", "push", imageName, fmt.Sprintf("docker://%s:latest", imageName)})

	return buildContainer.Stdout(ctx)
}

// BuildWithBuildkit alternativ med BuildKit
func (m *CiPipeline) BuildWithBuildkit(ctx context.Context, source *dagger.Directory, imageName string) (string, error) {
	if imageName == "" {
		imageName = "myapp"
	}

	// Kör tester först
	testContainer := dag.Container().
		From("golang:1.21-alpine").
		WithDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{"go", "mod", "download"}).
		WithExec([]string{"go", "test", "-v", "./..."})

	// Använd BuildKit för att bygga
	buildContainer := dag.Container().
		From("moby/buildkit:latest").
		WithDirectory("/workspace", source).
		WithExec([]string{
			"buildctl", "build",
			"--frontend", "dockerfile.v0",
			"--local", "context=/workspace",
			"--local", "dockerfile=/workspace",
			"--output", fmt.Sprintf("type=image,name=%s:latest,push=true", imageName),
		})

	return buildContainer.Stdout(ctx)
}
