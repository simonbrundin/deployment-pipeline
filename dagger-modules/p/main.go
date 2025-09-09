package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"dagger.io/dagger"
)

type Pipeline struct{}

// CI kör komplett CI-workflow
func (m *Pipeline) CI(projectFolder string) *dagger.Container {
	// Använd Buildkit som byggmotor istället för Docker. Buildkit kan till
	// skillnad från Docker bygga både lokalt och i Kubernetes.
	os.Setenv("DAGGER_ENGINE_BACKEND", "buildkit")
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		log.Fatalf("Failed to connect to Dagger: %v", err)
	}
	defer client.Close()

	startTotal := time.Now()
	// 1. Kör unit tests
	containerWithTests := m.UnitTests(projectFolder)

	// 2. Bygg image
	image := m.BuildImage(containerWithTests)

	// 3. Pusha image till registry
	final := m.PushImage(image)
	fmt.Printf("✅ Workflow completed successfully! Total time: %v\n", time.Since(startTotal))

	return final
}
