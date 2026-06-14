package main

import (
	"context"
	"dagger/pipeline/internal/dagger"
	"fmt"
	"time"
)

// MultiArchContainers håller containers för olika plattformar
type MultiArchContainers struct {
	Containers []*dagger.Container
	Platforms  []dagger.Platform
}

// BuildMultiArchImage bygger containers för flera arkitekturer utan att pusha
func (pipeline *Pipeline) BuildMultiArchImage(ctx context.Context, sourceDir *dagger.Directory) (*MultiArchContainers, error) {
	start := time.Now()
	logs := "📦 Bygger multi-arch containers...\n"

	// Plattformar att bygga för
	platforms := []dagger.Platform{
		"linux/amd64", // x86_64
		"linux/arm64", // aarch64
	}

	// Kolla om Dockerfile finns
	dockerfileExists := false
	if _, err := sourceDir.File("Dockerfile").Contents(ctx); err == nil {
		dockerfileExists = true
	}

	var containers []*dagger.Container

	if dockerfileExists {
		logs += "📄 Dockerfile hittad, bygger multi-arch från Dockerfile...\n"

		// Bygg för varje plattform med native emulation
		for _, platform := range platforms {
			logs += fmt.Sprintf("🔨 Bygger för %s...\n", platform)

			container := sourceDir.DockerBuild(dagger.DirectoryDockerBuildOpts{
				Platform: platform,
			})

			containers = append(containers, container)
		}
	} else {
		logs += "📦 Ingen Dockerfile, bygger Go-binärer med cross-compilation...\n"

		// Cross-compilation för Go-projekt
		for _, platform := range platforms {
			logs += fmt.Sprintf("🔨 Cross-kompilerar för %s...\n", platform)

			// Extrahera arkitektur från platform string
			var goarch string
			switch platform {
			case "linux/amd64":
				goarch = "amd64"
			case "linux/arm64":
				goarch = "arm64"
			default:
				goarch = "amd64" // fallback
			}

			// Bygg Go-binär med cross-compilation på host-plattformen
			builder := dag.Container().
				From("golang:1.21-alpine").
				WithDirectory("/src", sourceDir).
				WithWorkdir("/src").
				WithEnvVariable("CGO_ENABLED", "0").
				WithEnvVariable("GOOS", "linux").
				WithEnvVariable("GOARCH", goarch).
				WithExec([]string{"go", "build", "-o", "/output/app"})

			// Hämta den byggda binären
			binary := builder.File("/output/app")

			// Skapa minimal container för target-plattformen
			container := dag.Container(dagger.ContainerOpts{Platform: platform}).
				From("alpine:latest").
				WithFile("/app", binary).
				WithEntrypoint([]string{"/app"})

			containers = append(containers, container)
		}
	}

	logs += fmt.Sprintf("✅ Multi-arch containers färdigbyggda! Körtid: %ds\n", int(time.Since(start).Seconds()))
	fmt.Print(logs)

	return &MultiArchContainers{
		Containers: containers,
		Platforms:  platforms,
	}, nil
}
