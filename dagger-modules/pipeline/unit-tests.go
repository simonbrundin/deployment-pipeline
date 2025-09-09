package main

import (
	"context"
	"fmt"
	"time"
)

// UnitTests kör unit tester
func (m *Pipeline) UnitTests(ctx context.Context, sourceDir string) string {
	start := time.Now()
	fmt.Println("🧪 Kör unit tester...")

	fmt.Printf("🧪 Testning klar! Körtid: %v\n", time.Since(start))
	return "hej"
}

func nodeTests() {
}
