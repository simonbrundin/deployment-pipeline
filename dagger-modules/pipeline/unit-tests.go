package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"dagger.io/dagger"
)

// UnitTests kör tester med Bun på valfri katalog
func (m *Pipeline) UnitTests(ctx context.Context, source string) error {
	fmt.Println("🧪 Running unit tests...")

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(log.Writer()))
	if err != nil {
		return fmt.Errorf("failed to connect to Dagger: %w", err)
	}
	defer client.Close()

	// Gör path absolut
	if !filepath.IsAbs(source) {
		absPath, err := filepath.Abs(source)
		if err != nil {
			return fmt.Errorf("failed to resolve absolute path: %w", err)
		}
		source = absPath
	}

	// Kontrollera att katalogen finns på hosten
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return fmt.Errorf("source path does not exist: %s", source)
	}

	// ⚡ Mountar host-katalogen i Dagger
	hostDir := client.Host().Directory(source)

	// Kolla om package.json finns
	if _, err := hostDir.File("package.json").Contents(ctx); err != nil {
		fmt.Println("ℹ️ No package.json found, skipping tests")
		return nil
	}

	container := client.Container().
		From("oven/bun:latest").
		WithWorkdir("/app").
		WithMountedDirectory("/app", hostDir).
		WithMountedCache("/root/.bun", client.CacheVolume("bun-cache")).
		WithMountedCache("/app/node_modules", client.CacheVolume("node-modules-cache")).
		WithExec([]string{"bun", "install"})

	testOutput, err := container.WithExec([]string{"bun", "test"}).Stdout(ctx)
	if err != nil {
		return fmt.Errorf("failed to run tests: %w", err)
	}

	fmt.Println("───── Test Output ─────")
	fmt.Println(testOutput)
	return nil
}
