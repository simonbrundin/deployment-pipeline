package main

import (
	"context"
	"dagger/pipeline/internal/dagger"
	"fmt"
	"strings"
)

func (pipeline *Pipeline) GetLatestTag(ctx context.Context, sourceDir *dagger.Directory) (string, error) {
	// Använd git-container med caching för snabbare tag-hämtning
	container := dag.Container().
		From("alpine/git").
		WithMountedDirectory("/src", sourceDir).
		WithWorkdir("/src").
		WithExec([]string{"sh", "-c", "git describe --tags --match 'frontend-*' --abbrev=0 2>/dev/null || echo 'v0.0.0'"})

	result, err := container.Sync(ctx)
	if err != nil {
		return "v0.0.0", nil // Fallback om git inte fungerar
	}

	stdout, err := result.Stdout(ctx)
	if err != nil {
		return "v0.0.0", nil
	}

	tag := strings.TrimSpace(stdout)
	tag = strings.TrimPrefix(tag, "frontend-")
	if tag == "" {
		return "v0.0.0", nil
	}
	return tag, nil
}

func (pipeline *Pipeline) GetCommitsSinceTag(ctx context.Context, sourceDir *dagger.Directory, tag string) (string, error) {
	// Använd git-container med caching för snabbare commit-logg
	container := dag.Container().
		From("alpine/git").
		WithMountedDirectory("/src", sourceDir).
		WithWorkdir("/src").
		WithExec([]string{"sh", "-c", fmt.Sprintf("git log 'frontend-%s'..HEAD --oneline 2>/dev/null || git log --oneline -10 2>/dev/null || echo ''", tag)})

	result, err := container.Sync(ctx)
	if err != nil {
		return "", nil // Fallback om git inte fungerar
	}

	stdout, err := result.Stdout(ctx)
	if err != nil {
		return "", nil
	}

	return strings.TrimSpace(stdout), nil
}
