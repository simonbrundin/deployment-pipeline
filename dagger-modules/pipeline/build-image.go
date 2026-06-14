package main

import (
	"context"
	"dagger/pipeline/internal/dagger"
	"fmt"
	"time"
)

// BuildImage bygger en Image från Dockerfile eller direkt från källkoden
// Använder Docker layer caching via BuildKit för snabbare builds
func (pipeline *Pipeline) BuildImage(ctx context.Context, sourceDir *dagger.Directory) (*dagger.Container, error) {
	start := time.Now()
	logs := "📦 Bygger image...\n"

	var container *dagger.Container

	// Kolla om Dockerfile finns
	dockerfileExists := false
	if _, err := sourceDir.File("Dockerfile").Contents(ctx); err == nil {
		dockerfileExists = true
	}

	if dockerfileExists {
		logs += "📄 Dockerfile hittad, bygger container med Docker layer caching...\n"
		// Bygg container från Dockerfile med Dagger
		// Använd cache volym för Docker layer cache - detta dramatiskt
		// snabbar upp rebuilds genom att återanvända oförändrade lager
		container = sourceDir.DockerBuild(dagger.DirectoryDockerBuildOpts{
			// BuildKit cache mount för snabbare lager-bygg
			BuildArgs: []dagger.BuildArg{
				{Name: "BUILDKIT_INLINE_CACHE", Value: "1"},
			},
		})
	} else {
		logs += "📦 Ingen Dockerfile, bygger standard container...\n"
		// Bygg container från källkod - använd en enkel base image
		// med caching av package manager om det finns
		container = buildContainerWithCaching(ctx, sourceDir)
	}

	logs += fmt.Sprintf("✅ Container färdigbyggd! Körtid: %ds\n", int(time.Since(start).Seconds()))
	return container, nil
}

// buildContainerWithCaching bygger container med intelligent caching baserat på projekttyp
func buildContainerWithCaching(ctx context.Context, sourceDir *dagger.Directory) *dagger.Container {
	base := dag.Container().
		From("alpine:latest").
		WithWorkdir("/app")

	// Detektera projekttyp och lägg till relevant caching
	if _, err := sourceDir.File("package.json").Contents(ctx); err == nil {
		// JavaScript/Node projekt
		base = base.
			WithMountedCache("/root/.npm", dag.CacheVolume("npm-cache")).
			WithMountedCache("/app/node_modules", dag.CacheVolume("node-modules-cache"))
	} else if _, err := sourceDir.File("go.mod").Contents(ctx); err == nil {
		// Go projekt
		base = base.
			WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod-cache")).
			WithMountedCache("/go/build-cache", dag.CacheVolume("go-build-cache"))
	} else if _, err := sourceDir.File("requirements.txt").Contents(ctx); err == nil {
		// Python projekt
		base = base.
			WithEnvVariable("PIP_CACHE_DIR", "/root/.cache/pip").
			WithMountedCache("/root/.cache/pip", dag.CacheVolume("pip-cache"))
	}

	return base.WithDirectory("/app", sourceDir)
}
