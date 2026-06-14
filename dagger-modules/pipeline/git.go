package main

import (
	"context"
	"dagger/pipeline/internal/dagger"
	"fmt"
	"strings"
)

func (pipeline *Pipeline) GetLatestTag(ctx context.Context, sourceDir *dagger.Directory) (string, error) {
	container := dag.Container().
		From("alpine:latest").
		WithMountedDirectory("/src", sourceDir).
		WithWorkdir("/src").
		WithExec([]string{"git", "fetch", "--tags"})

	result, err := container.WithExec([]string{"sh", "-c", "git describe --tags --match 'frontend-*' --abbrev=0 2>/dev/null || echo 'v0.0.0'"}).Sync(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get latest tag: %w", err)
	}

	stdout, err := result.Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read stdout: %w", err)
	}

	tag := strings.TrimSpace(stdout)
	tag = strings.TrimPrefix(tag, "frontend-")
	return tag, nil
}

func (pipeline *Pipeline) GetCommitsSinceTag(ctx context.Context, sourceDir *dagger.Directory, tag string) (string, error) {
	container := dag.Container().
		From("alpine:latest").
		WithMountedDirectory("/src", sourceDir).
		WithWorkdir("/src").
		WithExec([]string{"git", "fetch", "--tags"})

	result, err := container.WithExec([]string{"sh", "-c", fmt.Sprintf("git log 'frontend-%s'..HEAD --oneline 2>/dev/null || git log --oneline -10", tag)}).Sync(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get commits: %w", err)
	}

	stdout, err := result.Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read stdout: %w", err)
	}

	return stdout, nil
}
