package main

import (
	"context"
	"dagger/pipeline/internal/dagger"
	"fmt"
	"time"
)

// UnitTests kör unit tester
func (pipeline *Pipeline) UnitTests(ctx context.Context, sourceDir *dagger.Directory) (string, error) {
	start := time.Now()
	logs := "🧪 Kör unit tester...\n"

	projectLanguage := detectProjectLanguage(ctx, sourceDir)
	switch projectLanguage {
	case "javascript":
		testLogs := javascriptTests(ctx, sourceDir)
		logs += testLogs
	case "go":
		testLogs, err := goTests(ctx, sourceDir)
		if err != nil {
			logs += fmt.Sprintf("❌ Fel vid körning av Go-tester: %v\n", err)
		} else {
			logs += testLogs
		}
	case "java":
		testLogs, err := javaTests(ctx, sourceDir)
		if err != nil {
			logs += fmt.Sprintf("❌ Fel vid körning av Java-tester: %v\n", err)
		} else {
			logs += testLogs
		}
	case "python":
		logs += "ℹ️ Python-tester inte implementerade ännu\n"
	default:
		logs += "ℹ️ Okänt projektspråk, hoppar över tester\n"
	}

	logs += fmt.Sprintf("✅ Tester klara! Körtid: %v\n", time.Since(start))
	return logs, nil
}

func javascriptTests(ctx context.Context, source *dagger.Directory) string {
	logs := "🧪 Kör JavaScript-tester...\n"

	// Grundcontainer med miljö och cache
	base := dag.Container().
		From("oven/bun:latest").
		WithWorkdir("/app").
		WithMountedDirectory("/app", source).
		WithMountedCache("/root/.bun", dag.CacheVolume("bun-cache")).
		WithMountedCache("/app/node_modules", dag.CacheVolume("node-modules-cache"))

	// Steg 1: installera beroenden
	deps := base.WithExec([]string{"bun", "install"})

	// Steg 2: kör tester
	deps.WithExec([]string{"bun", "test"}).Sync(ctx)
	logs += "✅ JavaScript-tester klara\n"

	return logs
}

func goTests(ctx context.Context, source *dagger.Directory) (string, error) {
	logs := "🧪 Kör Go-tester...\n"

	// Grundcontainer med Go-miljö
	base := dag.Container().
		From("golang:1.23"). // Uppdaterad till en nyare Go-version
		WithWorkdir("/app").
		WithMountedDirectory("/app", source).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod-cache")).
		WithMountedCache("/go/build-cache", dag.CacheVolume("go-build-cache"))

	// Steg 1: hämta beroenden
	deps, err := base.WithExec([]string{"go", "mod", "download"}).Sync(ctx)
	if err != nil {
		logs += fmt.Sprintf("❌ Fel vid hämtning av Go-beroenden: %v\n", err)
		return logs, err
	}

	// Steg 2: kör tester
	_, err = deps.WithExec([]string{"go", "test", "./...", "-v"}).Sync(ctx)
	if err != nil {
		logs += fmt.Sprintf("❌ Fel vid körning av Go-tester: %v\n", err)
		return logs, err
	}
	logs += "✅ Go-tester klara\n"

	return logs, nil
}

func javaTests(ctx context.Context, source *dagger.Directory) (string, error) {
	logs := "🧪 Kör Java-tester...\n"

	// Grundcontainer med Java-miljö
	base := dag.Container().
		From("maven:3.9-openjdk-21").
		WithWorkdir("/app").
		WithMountedDirectory("/app", source).
		WithMountedCache("/root/.m2", dag.CacheVolume("maven-cache"))

	// Steg 1: kör tester med Maven
	_, err := base.WithExec([]string{"mvn", "test"}).Sync(ctx)
	if err != nil {
		logs += fmt.Sprintf("❌ Fel vid körning av Java-tester: %v\n", err)
		return logs, err
	}
	logs += "✅ Java-tester klara\n"

	return logs, nil
}

func fileExists(ctx context.Context, sourceDir *dagger.Directory, fileName string) bool {
	_, err := sourceDir.File(fileName).Contents(ctx)
	return err == nil
}

func detectProjectLanguage(ctx context.Context, sourceDir *dagger.Directory) string {
	// Kolla efter typiska filer för olika projekt
	if fileExists(ctx, sourceDir, "package.json") {
		return "javascript"
	}
	if fileExists(ctx, sourceDir, "go.mod") {
		return "go"
	}
	if fileExists(ctx, sourceDir, "pom.xml") {
		return "java"
	}
	if fileExists(ctx, sourceDir, "pyproject.toml") || fileExists(ctx, sourceDir, "requirements.txt") {
		return "python"
	}
	return "unknown"
}
