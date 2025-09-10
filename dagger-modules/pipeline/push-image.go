package main

import (
	"context"
	"dagger/pipeline/internal/dagger"
	"fmt"
	"time"
)

// PushImage pushar image till registry
func (pipeline *Pipeline) PushImage(ctx context.Context, container *dagger.Container, registryAddress string, imageName string, tag string, username string, secret string) (string, error) {
	start := time.Now()
	logs := fmt.Sprintf("📤 Pushar image %s:%s till %s...\n", imageName, tag, registryAddress)

	fullImageName := fmt.Sprintf("%s/%s:%s", registryAddress, imageName, tag)

	// Publicera containern till registry med autentisering
	authContainer := container.WithRegistryAuth("ghcr.io", username, dag.SetSecret("password", secret))

	_, err := authContainer.Publish(ctx, fullImageName)
	if err != nil {
		logs += fmt.Sprintf("❌ Fel vid push av image: %v\n", err)
		return logs, err
	}

	logs += fmt.Sprintf("✅ Uppladdning färdig! Körtid: %ds\n", int(time.Since(start).Seconds()))
	return logs, nil
}
