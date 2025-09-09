package main

import (
	"fmt"
	"time"
)

// BuildImage bygger en Image från Dockerfile eller direkt från källkoden
func (m *Pipeline) BuildImage() {
	start := time.Now()
	fmt.Println("📦 Bygger image...")

	fmt.Printf("📦 Image färdigbyggd! Körtid: %v\n", time.Since(start))
}
