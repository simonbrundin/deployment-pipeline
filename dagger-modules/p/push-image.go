package main

import "dagger.io/dagger"

// PushImage pushar image till registry
func (m *Pipeline) PushImage(container *dagger.Container) *dagger.Container {
	return container.
		WithExec([]string{"sh", "-c", "docker push my-image:latest"})
}
