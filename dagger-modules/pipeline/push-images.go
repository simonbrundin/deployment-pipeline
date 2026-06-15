package main

import (
	"context"
	"dagger/pipeline/internal/dagger"
	"fmt"
	"os"
	"time"
)

// PushImages pushar en eller flera container-images till registry
func (pipeline *Pipeline) PushImages(
	ctx context.Context,
	containers []*dagger.Container,
	registryAddress string,
	imageName string,
	tag string,
	username string,
) (string, error) {
	start := time.Now()
	archType := "single"
	if len(containers) > 1 {
		archType = "multi-arch"
	}

	logs := fmt.Sprintf("📤 Pushar %s image %s:%s till %s...\n", archType, imageName, tag, registryAddress)
	fullImageName := fmt.Sprintf("%s/%s:%s", registryAddress, imageName, tag)

	// Läs token från GITHUB_TOKEN (GitHub Actions built-in) eller REGISTRY_PASSWORD
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("REGISTRY_PASSWORD")
	}
	secret := dag.SetSecret("password", token)

	// Lägg till autentisering för alla containers
	var authContainers []*dagger.Container
	for _, container := range containers {
		authContainer := container.WithRegistryAuth(registryAddress, username, secret)
		authContainers = append(authContainers, authContainer)
	}

	// Publicera image (single eller multi-arch)
	_, err := dag.Container().
		Publish(ctx, fullImageName, dagger.ContainerPublishOpts{
			PlatformVariants: authContainers,
		})
	if err != nil {
		logs += fmt.Sprintf("❌ Push misslyckades: %v\n", err)
		return logs, err
	}

	logs += fmt.Sprintf("✅ Push klar! Tid: %ds\n", int(time.Since(start).Seconds()))
	return logs, nil
}
