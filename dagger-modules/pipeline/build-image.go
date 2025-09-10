package main

import (
	"context"
	"dagger/pipeline/internal/dagger"
	"fmt"
	"time"
)

// BuildImage bygger en Image från Dockerfile eller direkt från källkoden
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
		logs += "📄 Dockerfile hittad, bygger container från Dockerfile...\n"
		// Bygg container från Dockerfile med Dagger
		container = dag.Container().Build(sourceDir)
	} else {
		logs += "📦 Ingen Dockerfile, bygger standard container...\n"
		// Bygg container från källkod - använd en enkel base image
		container = dag.Container().
			From("alpine:latest").
			WithWorkdir("/app").
			WithDirectory("/app", sourceDir)
	}

	logs += fmt.Sprintf("✅ Container färdigbyggd! Körtid: %v\n", time.Since(start))
	return container, nil
}
