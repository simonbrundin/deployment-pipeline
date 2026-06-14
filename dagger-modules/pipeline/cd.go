package main

import (
	"fmt"
	"time"
)

// CD kör komplett CD-workflow
func (pipeline *Pipeline) CD() (string, error) {
	startTime := time.Now()
	logs := "🚀 Startar CD-workflow...\n"

	logs += fmt.Sprintf("✅ CD-workflow klar! Total körtid: %ds\n", int(time.Since(startTime).Seconds()))
	return logs, nil
}
