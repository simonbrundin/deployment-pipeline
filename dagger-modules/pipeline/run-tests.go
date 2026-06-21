package main

import (
	"context"
	"fmt"
	"time"

	"dagger/pipeline/internal/dagger"
)

// RunTests kör tester via cd ./tests && bun run test
func (pipeline *Pipeline) RunTests(ctx context.Context, sourceDir *dagger.Directory) (string, error) {
	start := time.Now()
	logs := "🧪 Kör tester...\n"

	// Kör tester i tests/-mappen
	container := dag.Container().
		From("oven/bun:latest").
		WithWorkdir("/app").
		WithMountedCache("/root/.bun", dag.CacheVolume("bun-cache")).
		WithDirectory("/app", sourceDir).
		WithExec([]string{"sh", "-c", "cd tests && bun test"})

	stdout, err := container.Stdout(ctx)
	if err != nil {
		logs += fmt.Sprintf("❌ Fel vid körning av tester: %v\n", err)
		return logs, err
	}

	logs += stdout
	logs += fmt.Sprintf("✅ Tester klara! Körtid: %ds\n", int(time.Since(start).Seconds()))
	return logs, nil
}
