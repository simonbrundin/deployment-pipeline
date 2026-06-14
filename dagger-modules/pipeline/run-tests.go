package main

import (
	"context"
	"dagger/pipeline/internal/dagger"
	"fmt"
	"strings"
	"time"
)

// RunTests kör tester med taggfilter
// tagFilter: cucumber-tag att filtrera på (t.ex. "@commit", "@ci", "not @commit"). Nil eller tom sträng = kör alla.
func (pipeline *Pipeline) RunTests(ctx context.Context, sourceDir *dagger.Directory, tagFilter *string) (string, error) {
	start := time.Now()
	logs := "🧪 Kör unit tester...\n"

	tagFilterValue := ""
	if tagFilter != nil {
		tagFilterValue = *tagFilter
	}

	if tagFilterValue != "" {
		logs += fmt.Sprintf("🏷️  Filtrerar på tagg: %s\n", tagFilterValue)
	}

	projectLanguage := detectProjectLanguage(ctx, sourceDir)
	switch projectLanguage {
	case "javascript":
		testLogs := javascriptTests(ctx, sourceDir, tagFilterValue)
		logs += testLogs
	case "go":
		testLogs, err := goTests(ctx, sourceDir, tagFilterValue)
		if err != nil {
			logs += fmt.Sprintf("❌ Fel vid körning av Go-tester: %v\n", err)
		} else {
			logs += testLogs
		}
	case "java":
		testLogs, err := javaTests(ctx, sourceDir, tagFilterValue)
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

	logs += fmt.Sprintf("✅ Tester klara! Körtid: %ds\n", int(time.Since(start).Seconds()))
	return logs, nil
}

func javascriptTests(ctx context.Context, source *dagger.Directory, tagFilter string) string {
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
	var testArgs []string
	if tagFilter != "" {
		// Hantera både inkludering (t.ex. "@commit") och exkludering (t.ex. "not @commit")
		if strings.HasPrefix(tagFilter, "not ") {
			testArgs = []string{"bun", "run", "test", "--tags", tagFilter}
		} else {
			testArgs = []string{"bun", "run", "test", "--tags", tagFilter}
		}
	} else {
		testArgs = []string{"bun", "run", "test"}
	}
	result := deps.WithExec(testArgs)
	stdout, err := result.Stdout(ctx)
	if err != nil {
		logs += fmt.Sprintf("❌ Fel vid körning av JavaScript-tester: %v\n", err)
	} else {
		logs += stdout
	}

	return logs
}

func goTests(ctx context.Context, source *dagger.Directory, tagFilter string) (string, error) {
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
	var testArgs []string
	if tagFilter != "" {
		testArgs = []string{"go", "test", "./...", "-v", "-run", tagFilter}
	} else {
		testArgs = []string{"go", "test", "./...", "-v"}
	}
	_, err = deps.WithExec(testArgs).Sync(ctx)
	if err != nil {
		logs += fmt.Sprintf("❌ Fel vid körning av Go-tester: %v\n", err)
		return logs, err
	}

	return logs, nil
}

func javaTests(ctx context.Context, source *dagger.Directory, tagFilter string) (string, error) {
	logs := "🧪 Kör Java-tester...\n"

	// Grundcontainer med Java-miljö
	base := dag.Container().
		From("maven:3.9-openjdk-21").
		WithWorkdir("/app").
		WithMountedDirectory("/app", source).
		WithMountedCache("/root/.m2", dag.CacheVolume("maven-cache"))

	// Steg 1: kör tester med Maven
	var testArgs []string
	if tagFilter != "" {
		testArgs = []string{"mvn", "test", "-Dcucumber.filter.tags=" + tagFilter}
	} else {
		testArgs = []string{"mvn", "test"}
	}
	_, err := base.WithExec(testArgs).Sync(ctx)
	if err != nil {
		logs += fmt.Sprintf("❌ Fel vid körning av Java-tester: %v\n", err)
		return logs, err
	}

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
