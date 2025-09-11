package main

import (
	"context"
	"dagger/pipeline/internal/dagger"
	"fmt"
	"time"
)

type Pipeline struct{}

// CI kör komplett CI-workflow
func (pipeline *Pipeline) CI(
	sourceDir *dagger.Directory,
	registryAddress string,
	imageName string,
	tag string,
	username string,
	secret string,
	multiArch bool, // 🔑 optional med default true
) (string, error) {
	startTime := time.Now()
	ctx := context.Background()
	logs := "🚀 Startar CI-workflow...\n"

	// Sätt default värde för multiArch
	useMultiArch := multiArch

	// 1. Kör unit tests
	testLogs, err := pipeline.UnitTests(ctx, sourceDir)
	if err != nil {
		logs += fmt.Sprintf("❌ Fel vid körning av tester: %v\n", err)
		return logs, err
	}
	logs += testLogs

	if useMultiArch {
		// 2a. Bygg multi-arch (default)
		multiArchContainers, err := pipeline.BuildMultiArchImage(ctx, sourceDir)
		if err != nil {
			logs += fmt.Sprintf("❌ Fel vid byggande av multi-arch image: %v\n", err)
			return logs, err
		}
		logs += "✅ Multi-arch containers byggda framgångsrikt\n"

		// 3a. Pusha multi-arch
		pushLogs, err := pipeline.PushMultiArchImage(ctx, multiArchContainers, registryAddress, imageName, tag, username, secret)
		if err != nil {
			logs += fmt.Sprintf("❌ Fel vid push av multi-arch image: %v\n", err)
			return logs, err
		}
		logs += pushLogs

	} else {
		// 2b. Bygg single-arch (endast om explicit false)
		container, err := pipeline.BuildImage(ctx, sourceDir)
		if err != nil {
			logs += fmt.Sprintf("❌ Fel vid byggande av image: %v\n", err)
			return logs, err
		}
		logs += "✅ Container byggd framgångsrikt\n"

		// 3b. Pusha single-arch
		pushLogs, err := pipeline.PushImage(ctx, container, registryAddress, imageName, tag, username, secret)
		if err != nil {
			logs += fmt.Sprintf("❌ Fel vid push av image: %v\n", err)
			return logs, err
		}
		logs += pushLogs
	}

	logs += fmt.Sprintf("✅ CI-workflow klar! Total körtid: %ds\n", int(time.Since(startTime).Seconds()))
	return logs, nil
}
