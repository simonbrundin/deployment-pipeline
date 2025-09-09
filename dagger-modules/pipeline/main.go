package main

import (
	"dagger/pipeline/internal/dagger"
)

type Pipeline struct{}

// CI kör fullt CI-workflow
func (m *Pipeline) CI(stringArg string) *dagger.Container {
	// 1. Kör unit tests
	containerWithTests := m.UnitTests(stringArg)

	// 2. Bygg image
	image := m.BuildImage(containerWithTests)

	// 3. Pusha image till registry
	final := m.PushImage(image)

	return final
}
