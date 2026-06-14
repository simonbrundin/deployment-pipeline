package main

import (
	"context"
	"dagger/pipeline/internal/dagger"
	"fmt"
	"time"
)

// CI är main-funktionen som kör hela CI-flödet:
// 1. Kör tester
// 2. Bygg container-image
// 3. Pusha till registry
func (pipeline *Pipeline) CI(
	sourceDir *dagger.Directory,
	registryAddress string,
	imageName string,
	tag string,
	username string,
	secret string,
	multiArch bool,
) (string, error) {
	// ---- FÖRBEREDelser ----
	startTime := time.Now()
	ctx := context.Background()
	logs := "🚀 Startar CI-workflow...\n"

	// ============================================
	// STEG 1: KÖR ENHETSTESTER
	// ============================================
	testLogs, err := pipeline.UnitTests(ctx, sourceDir)
	if err != nil {
		return logs + fmt.Sprintf("❌ Test misslyckades: %v\n", err), err
	}
	logs += testLogs

	// ============================================
	// STEG 2: BYGG CONTAINER-IMAGE
	// ============================================
	var containers []*dagger.Container

	if multiArch {
		// Bygger imagen för MULTIPLA arkitekturer (amd64 + arm64)
		result, err := pipeline.BuildMultiArchImage(ctx, sourceDir)
		if err != nil {
			return logs + fmt.Sprintf("❌ Bygge misslyckades: %v\n", err), err
		}
		containers = result.Containers
		logs += "✅ Image byggd (multi-arch)\n"

	} else {
		// Bygger imagen för EN arkitektur (snabbare)
		container, err := pipeline.BuildImage(ctx, sourceDir)
		if err != nil {
			return logs + fmt.Sprintf("❌ Bygge misslyckades: %v\n", err), err
		}
		containers = []*dagger.Container{container}
		logs += "✅ Image byggd (single-arch)\n"
	}

	// ============================================
	// STEG 3: PUSHA TILL REGISTRY (GEMENSAMT!)
	// ============================================
	pushLogs, err := pipeline.PushImages(ctx, containers, registryAddress, imageName, tag, username, secret)
	if err != nil {
		return logs + fmt.Sprintf("❌ Push misslyckades: %v\n", err), err
	}
	logs += pushLogs

	// ---- KLART! ----
	logs += fmt.Sprintf("✅ CI klart! Tid: %ds\n", int(time.Since(startTime).Seconds()))
	return logs, nil
}
