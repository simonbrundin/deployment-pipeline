// package main
//
// import (
// 	"context"
// 	"dagger/pipeline/internal/dagger"
// 	"fmt"
// 	"time"
// )
//
// type Pipeline struct{}
//
// // CI kör komplett CI-workflow
// func (pipeline *Pipeline) CI(sourceDir *dagger.Directory, registryAddress string, imageName string, tag string, username string, secret string) (string, error) {
// 	startTime := time.Now()
// 	ctx := context.Background()
// 	logs := "🚀 Startar CI-workflow...\n"
//
// 	// 1. Kör unit tests
// 	testLogs, err := pipeline.UnitTests(ctx, sourceDir)
// 	if err != nil {
// 		logs += fmt.Sprintf("❌ Fel vid körning av tester: %v\n", err)
// 		return logs, err
// 	}
// 	logs += testLogs
//
// 	// 2. Bygg image
// 	container, err := pipeline.BuildImage(ctx, sourceDir)
// 	if err != nil {
// 		logs += fmt.Sprintf("❌ Fel vid byggande av image: %v\n", err)
// 		return logs, err
// 	}
// 	logs += "✅ Container byggd framgångsrikt\n"
//
// 	// 3. Pusha image till registry
// 	pushLogs, err := pipeline.PushImage(ctx, container, registryAddress, imageName, tag, username, secret)
// 	if err != nil {
// 		logs += fmt.Sprintf("❌ Fel vid push av image: %v\n", err)
// 		return logs, err
// 	}
// 	logs += pushLogs
//
// 	logs += fmt.Sprintf("✅ CI-workflow klar! Total körtid: %ds\n", int(time.Since(startTime).Seconds()))
// 	return logs, nil
// }
