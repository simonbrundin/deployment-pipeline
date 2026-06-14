package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SemVerBump analyzes a commit message and calculates the next semantic version
// based on Conventional Commits specification.
//
// Rules:
//   - "BREAKING CHANGE:" in commit message → major bump
//   - "feat:" in commit message → minor bump
//   - "fix:" or other → patch bump
//   - No match → patch bump (default)
func (pipeline *Pipeline) SemVerBump(
	currentVersion string,
	commitMessage string,
) (string, error) {
	bumpType := determineBumpType(commitMessage)
	newVersion, err := incrementVersion(currentVersion, bumpType)
	if err != nil {
		return "", fmt.Errorf("failed to increment version: %w", err)
	}

	return newVersion, nil
}

// determineBumpType analyzes the commit message to determine the bump type
func determineBumpType(message string) string {
	msg := strings.ToLower(message)

	// Check for breaking changes first (highest priority)
	if strings.Contains(msg, "breaking change") || strings.Contains(msg, "breaking-change") {
		return "major"
	}

	// Check for feat (new feature = minor bump)
	featRegex := regexp.MustCompile(`^feat(\(.+\))?:`)
	if featRegex.MatchString(message) {
		return "minor"
	}

	// Default to patch bump
	return "patch"
}

// incrementVersion increments the version based on bump type
func incrementVersion(version string, bumpType string) (string, error) {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid version format: %s (expected semver)", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid minor version: %s", parts[1])
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", fmt.Errorf("invalid patch version: %s", parts[2])
	}

	switch bumpType {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	default:
		patch++
	}

	return fmt.Sprintf("v%d.%d.%d", major, minor, patch), nil
}

