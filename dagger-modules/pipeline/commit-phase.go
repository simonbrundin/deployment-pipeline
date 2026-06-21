package main

import (
	"context"
	"dagger/pipeline/internal/dagger"
	"fmt"
	"strings"
	"time"
)

// CI är main-funktionen som kör hela CI-flödet:
// 1. Kör tester
// 2. Bygg container-image
// 3. Pusha till registry
func (pipeline *Pipeline) CommitPhase(
	sourceDir *dagger.Directory,
	registryAddress string,
	imageName string,
	username string,
	multiArch bool,
) (string, error) {
	// ---- FÖRBEREDelser ----
	startTime := time.Now()
	ctx := context.Background()
	logs := "🚀 Startar CI-workflow...\n"

	// ============================================
	// STEG 0: BERÄKNA VERSION FRÅN GIT
	// ============================================
	latestTag, err := pipeline.GetLatestTag(ctx, sourceDir)
	if err != nil {
		logs += fmt.Sprintf("⚠️ Kunde inte hämta senaste tagg: %v\n", err)
		latestTag = "0.0.0"
	}
	logs += fmt.Sprintf("📌 Senaste tagg: %s\n", latestTag)

	commits, err := pipeline.GetCommitsSinceTag(ctx, sourceDir, latestTag)
	if err != nil {
		logs += fmt.Sprintf("⚠️ Kunde inte hämta commits: %v\n", err)
		commits = ""
	}

	commitMessage := "fix: update"
	if commits != "" {
		lines := strings.Split(strings.TrimSpace(commits), "\n")
		if len(lines) > 0 {
			commitMessage = strings.TrimSpace(lines[0])
			if idx := strings.Index(commitMessage, " "); idx > 0 {
				commitMessage = commitMessage[idx+1:]
			}
		}
	}
	logs += fmt.Sprintf("📝 Senaste commit: %s\n", commitMessage)

	newVersion, err := pipeline.SemVerBump(latestTag, commitMessage)
	if err != nil {
		logs += fmt.Sprintf("⚠️ Kunde inte beräkna ny version: %v\n", err)
		newVersion = "v1.0.0"
	}
	logs += fmt.Sprintf("🏷️  Ny version: %s\n", newVersion)

	tag := fmt.Sprintf("frontend-%s", newVersion)

	// ============================================
	// STEG 1: KÖR ENHETSTESTER
	// ============================================
	testLogs, err := pipeline.RunTests(ctx)
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
	pushLogs, err := pipeline.PushImages(ctx, containers, registryAddress, imageName, tag, username)
	if err != nil {
		return logs + fmt.Sprintf("❌ Push misslyckades: %v\n", err), err
	}
	logs += pushLogs

	// ---- KLART! ----
	logs += fmt.Sprintf("✅ CI klart! Tid: %ds\n", int(time.Since(startTime).Seconds()))
	return logs, nil
}
