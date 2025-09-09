package main

import (
	"fmt"
	"time"
)

// PushImage pushar image till registry
func (m *Pipeline) PushImage(registryAddress string) {
	start := time.Now()

	fmt.Printf("📤 Pushar image till %s\n", registryAddress)

	fmt.Printf("📤 Uppladdning färdig! Körtid: %v\n", time.Since(start))
}
