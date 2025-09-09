package main

import (
	"dagger/pipeline/internal/dagger"
)

// UnitTests kör unit tester
func (m *Pipeline) UnitTests(arg string) *dagger.Container {
	return dag.Container().
		From("golang:1.21-alpine").
		WithWorkdir("/app").
		WithExec([]string{"sh", "-c", "go test ./..."}).
		WithExec([]string{"echo", "Tests complete for " + arg})
}
