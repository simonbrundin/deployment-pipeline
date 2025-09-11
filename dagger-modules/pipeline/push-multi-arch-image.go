package main

import (
	"context"
	"dagger/pipeline/internal/dagger"
	"fmt"
	"time"
)

// PushMultiArchImage pushar multi-arch containers till registry
func (pipeline *Pipeline) PushMultiArchImage(ctx context.Context, multiArch *MultiArchContainers, registryAddress string, imageName string, tag string, username string, secret string) (string, error) {
	start := time.Now()
	logs := fmt.Sprintf("📤 Pushar multi-arch image %s:%s till %s...\n", imageName, tag, registryAddress)

	fullImageName := fmt.Sprintf("%s/%s:%s", registryAddress, imageName, tag)

	// Lägg till autentisering för alla containers
	var authContainers []*dagger.Container
	for _, container := range multiArch.Containers {
		authContainer := container.WithRegistryAuth(registryAddress, username, dag.SetSecret("password", secret))
		authContainers = append(authContainers, authContainer)
	}

	// Publicera multi-arch manifest
	_, err := dag.Container().
		Publish(ctx, fullImageName, dagger.ContainerPublishOpts{
			PlatformVariants: authContainers,
		})
	if err != nil {
		logs += fmt.Sprintf("❌ Fel vid push av multi-arch image: %v\n", err)
		return logs, err
	}

	logs += fmt.Sprintf("✅ Multi-arch uppladdning färdig! Körtid: %ds\n", int(time.Since(start).Seconds()))
	return logs, nil
}
