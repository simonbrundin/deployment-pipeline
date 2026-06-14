package main

import (
	"context"
	"dagger/pipeline/internal/dagger"
	"fmt"
	"time"
)

// AcceptancePhase kör acceptance-tester (tester utan @commit-tagg)
// och kan köras separat från CI-flödet för manuell verifiering.
func (pipeline *Pipeline) AcceptancePhase(
	ctx context.Context,
	sourceDir *dagger.Directory,
) (string, error) {
	startTime := time.Now()
	logs := "🚀 Startar Acceptance-testworkflow...\n"

	// Kör tester MEDtaggen "not @commit" för att exkludera commit-tester
	testLogs, err := pipeline.RunTests(ctx, sourceDir, "not @commit")
	if err != nil {
		return logs + fmt.Sprintf("❌ Acceptance-tester misslyckades: %v\n", err), err
	}
	logs += testLogs

	// ---- KLART! ----
	logs += fmt.Sprintf("✅ Acceptance-tester klara! Tid: %ds\n", int(time.Since(startTime).Seconds()))
	return logs, nil
}
