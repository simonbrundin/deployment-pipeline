package main

import "dagger/pipeline/internal/dagger"

// BuildImage bygger en Image från Dockerfile eller direkt från källkoden
func (m *Pipeline) BuildImage(container *dagger.Container) *dagger.Container {
	return container.
		WithExec([]string{"sh", "-c", "docker build -t my-image:latest ."})
}
