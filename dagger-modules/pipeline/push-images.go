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

	// Läs token från miljövariabel
	token := os.Getenv("REGISTRY_PASSWORD")
	if token == "" {
		return logs + "❌ REGISTRY_PASSWORD miljövariabel saknas\n", fmt.Errorf("REGISTRY_PASSWORD is required")
	}
	secret := dag.SetSecret("registry-password", token)

	// Lägg till autentisering och pusha direkt på varje container
	for i, container := range containers {
		authContainer := container.WithRegistryAuth(registryAddress, username, secret)
		// För multi-arch, använd första containern för push med PlatformVariants
		if i == 0 {
			var allAuthContainers []*dagger.Container
			for _, c := range containers {
				allAuthContainers = append(allAuthContainers, c.WithRegistryAuth(registryAddress, username, secret))
			}
			_, err := authContainer.Publish(ctx, fullImageName, dagger.ContainerPublishOpts{
				PlatformVariants: allAuthContainers,
			})
			if err != nil {
				logs += fmt.Sprintf("❌ Push misslyckades: %v\n", err)
				return logs, err
			}
			break
		}
	}

	logs += fmt.Sprintf("✅ Push klar! Tid: %ds\n", int(time.Since(start).Seconds()))
	return logs, nil
}
