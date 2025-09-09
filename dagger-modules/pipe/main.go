package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"dagger.io/dagger"
)

type Pipeline struct{}

func Ci() {
	ctx := context.Background()

	// Sätt miljövariabel för att använda BuildKit backend
	os.Setenv("DAGGER_ENGINE_BACKEND", "buildkit")

	sourceDir := flag.String("source", ".", "Path to project source directory")
	registryAddr := flag.String("registry", "", "Registry address to push to (optional)")
	username := flag.String("username", "", "Registry username")
	password := flag.String("password", "", "Registry password (use env:VAR_NAME for secrets)")
	flag.Parse()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		log.Fatalf("Failed to connect to Dagger: %v", err)
	}
	defer client.Close()

	startTotal := time.Now()

	if err := runWorkflow(ctx, client, *sourceDir, *registryAddr, *username, *password); err != nil {
		log.Fatalf("Workflow failed: %v", err)
	}

	fmt.Printf("✅ Workflow completed successfully! Total time: %v\n", time.Since(startTotal))
}

func runWorkflow(ctx context.context, client *dagger.client, sourcedir, registryaddr, username, password string) error {
	source := client.Host().Directory(sourceDir)

	// 1. Kör tester
	fmt.Println("🧪 Running tests...")
	start := time.Now()
	if err := runTests(ctx, client, source); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}
	fmt.Printf("🧪 Tests completed in %v\n", time.Since(start))

	// 2. Bygg image
	fmt.Println("🏗️ Building image...")
	start = time.Now()
	image, err := buildImage(ctx, client, source)
	if err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}
	fmt.Printf("🏗️ Image build completed in %v\n", time.Since(start))

	// 3. Pusha till registry om adress är angiven
	if registryAddr != "" {
		fmt.Printf("📤 Publishing to %s...\n", registryAddr)

		if username == "" || password == "" {
			return fmt.Errorf("username and password are required for registry push")
		}

		// Hantera secrets från miljövariabler (env:VAR_NAME format)
		var passwordSecret *dagger.Secret
		if len(password) > 4 && password[:4] == "env:" {
			envVar := password[4:]
			envValue := os.Getenv(envVar)
			if envValue == "" {
				return fmt.Errorf("environment variable %s is not set", envVar)
			}
			passwordSecret = client.SetSecret("registry-password", envValue)
		} else {
			passwordSecret = client.SetSecret("registry-password", password)
		}

		addr, err := image.
			WithRegistryAuth("ghcr.io", username, passwordSecret).
			Publish(ctx, registryAddr)
		if err != nil {
			return fmt.Errorf("failed to publish image: %w", err)
		}
		fmt.Printf("✅ Published image: %s\n", addr)
	} else {
		fmt.Println("ℹ️ No registry specified, skipping publish")
	}

	return nil
}

func runTests(ctx context.Context, client *dagger.Client, source *dagger.Directory) error {
	if _, err := source.File("package.json").Contents(ctx); err != nil {
		fmt.Println("ℹ️ No package.json found, skipping tests")
		return nil
	}

	baseContainer := client.Container().
		From("oven/bun:latest").
		WithWorkdir("/app").
		WithMountedDirectory("/app", source).
		WithMountedCache("/root/.bun", client.CacheVolume("bun-cache")).
		WithMountedCache("/app/node_modules", client.CacheVolume("node-modules-cache"))

	container := baseContainer.WithExec([]string{"bun", "install"})

	bunInstallOutput, err := container.Stdout(ctx)
	if err != nil {
		return fmt.Errorf("failed to run bun install: %w", err)
	}
	fmt.Println("───── bun install output ─────")
	fmt.Println(bunInstallOutput)

	container = container.WithExec([]string{"bun", "test"})

	testOutput, err := container.Stdout(ctx)
	if err != nil {
		return fmt.Errorf("failed to run tests: %w", err)
	}
	fmt.Println("───── Test Output ─────")
	fmt.Println(testOutput)

	return nil
}

func buildImage(ctx context.Context, client *dagger.Client, source *dagger.Directory) (*dagger.Container, error) {
	if _, err := source.File("Dockerfile").Contents(ctx); err == nil {
		fmt.Println("🐳 Building with Dockerfile with enhanced Docker layer caching")
		return client.Container().Build(source, dagger.ContainerBuildOpts{
			Dockerfile: "Dockerfile",
			BuildArgs: []dagger.BuildArg{{
				Name:  "BUILDKIT_INLINE_CACHE",
				Value: "1",
			}},
		}), nil
	}

	if _, err := source.File("package.json").Contents(ctx); err == nil {
		fmt.Println("📦 No Dockerfile found, building Node.js container with bun and cache")

		baseContainer := client.Container().
			From("oven/bun:latest").
			WithMountedCache("/root/.bun", client.CacheVolume("bun-cache")).
			WithMountedCache("/app/node_modules", client.CacheVolume("node-modules-cache")).
			WithMountedDirectory("/app", source).
			WithWorkdir("/app")

		container := baseContainer.
			WithExec([]string{"bun", "install"}).
			WithExposedPort(3000).
			WithEntrypoint([]string{"bun", "run", "start"})

		return container, nil
	}

	if _, err := source.File("go.mod").Contents(ctx); err == nil {
		fmt.Println("🐹 No Dockerfile found, building Go container")
		return client.Container().
			From("golang:alpine").
			WithMountedCache("/go/pkg/mod", client.CacheVolume("go-mod-cache")).
			WithMountedCache("/root/.cache/go-build", client.CacheVolume("go-build-cache")).
			WithDirectory("/app", source).
			WithWorkdir("/app").
			WithExec([]string{"go", "mod", "download"}).
			WithExec([]string{"go", "build", "-o", "main", "."}).
			WithEntrypoint([]string{"./main"}), nil
	}

	return nil, fmt.Errorf("no Dockerfile, package.json, or go.mod found - cannot determine how to build")
}
