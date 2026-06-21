package main

import (
	"context"
	"fmt"
	"time"

	"dagger/pipeline/internal/dagger"
)

// RunTests kör tester via bun test (från frontend-mappen)
func (pipeline *Pipeline) RunTests(ctx context.Context, sourceDir *dagger.Directory) (string, error) {
	start := time.Now()
	logs := "🧪 Kör tester...\n"

	// Kör tester (sourceDir innehåller frontend/)
	container := dag.Container().
		From("oven/bun:latest").
		WithWorkdir("/app").
		WithMountedCache("/root/.bun", dag.CacheVolume("bun-cache")).
		WithDirectory("/app", sourceDir).
		WithExec([]string{"bun", "test"})

	stdout, err := container.Stdout(ctx)
	if err != nil {
		logs += fmt.Sprintf("❌ Fel vid körning av tester: %v\n", err)
		return logs, err
	}

	logs += stdout
	logs += fmt.Sprintf("✅ Tester klara! Körtid: %ds\n", int(time.Since(start).Seconds()))
	return logs, nil
}
