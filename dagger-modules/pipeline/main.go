package main

import (
	"dagger/pipeline/internal/dagger"
	"os"
)

type Pipeline struct{}

// CI kör komplett CI-workflow
func (m *Pipeline) CI(stringArg string) *dagger.Container {
	// Använd Buildkit som byggmotor istället för Docker. Buildkit kan till
	// skillnad från Docker bygga både lokalt och i Kubernetes.
	os.Setenv("DAGGER_ENGINE_BACKEND", "buildkit")

	// 1. Kör unit tests
	containerWithTests := m.UnitTests(stringArg)

	// 2. Bygg image
	image := m.BuildImage(containerWithTests)

	// 3. Pusha image till registry
	final := m.PushImage(image)

	return final
}
