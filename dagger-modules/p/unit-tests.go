package main

import (
	"dagger.io/dagger"
	"dagger.io/dagger/dag"
)

// UnitTests kör unit tester
func (m *Pipeline) UnitTests(sourceDir string) *dagger.Container {
	return dag.Container().
		From("golang:1.21-alpine").
		WithWorkdir("/app").
		WithExec([]string{"sh", "-c", "go test ./..."}).
		WithExec([]string{"echo", "Tests complete for " + sourceDir})
}
