package main

import (
	"context"
	"fmt"
	"time"
)

// AcceptancePhase kör acceptance-tester
// och kan köras separat från CI-flödet för manuell verifiering.
func (pipeline *Pipeline) AcceptancePhase(
	ctx context.Context,
) (string, error) {
	startTime := time.Now()
	logs := "🚀 Startar Acceptance-testworkflow...\n"

	// Kör tester
	testLogs, err := pipeline.RunTests(ctx)
	if err != nil {
		return logs + fmt.Sprintf("❌ Acceptance-tester misslyckades: %v\n", err), err
	}
	logs += testLogs

	// ---- KLART! ----
	logs += fmt.Sprintf("✅ Acceptance-tester klara! Tid: %ds\n", int(time.Since(startTime).Seconds()))
	return logs, nil
}
