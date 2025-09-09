package main

import (
	"context"
	"fmt"
	"os"
	"time"
)

type Pipeline struct{}

// CI kör komplett CI-workflow
func (m *Pipeline) CI(projectFolder string, registryAddress string) {
	startTotal := time.Now()    // För att mäta hur lång tid allt tar
	ctx := context.Background() // Context för att styra och avbryta Dagger-operationer
	// Använd Buildkit som byggmotor istället för Docker. Buildkit kan till
	// skillnad från Docker bygga både lokalt och i Kubernetes.
	os.Setenv("DAGGER_ENGINE_BACKEND", "buildkit")

	// 1. Kör unit tests
	m.UnitTests(ctx, projectFolder)

	// 2. Bygg image
	m.BuildImage()

	// 3. Pusha image till registry
	m.PushImage(registryAddress)

	fmt.Printf("✅ CI-workflow lyckades! Körtid: %v\n", time.Since(startTotal))
}
