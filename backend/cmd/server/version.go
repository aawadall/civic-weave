package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Version information
var (
	Version   = getVersionFromFile()
	BuildTime = time.Now().Format("2006-01-02T15:04:05Z")
	GitCommit = getEnvOrDefault("GIT_COMMIT", "unknown")
	BuildEnv  = getEnvOrDefault("BUILD_ENV", "development")
)

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getVersionFromFile reads version from VERSION file or returns default
func getVersionFromFile() string {
	// First check environment variable (for Docker builds)
	if version := os.Getenv("VERSION"); version != "" {
		return version
	}
	
	// Try to read from VERSION file
	if data, err := os.ReadFile("VERSION"); err == nil {
		version := strings.TrimSpace(string(data))
		if version != "" {
			return version
		}
	}
	
	// Fallback to default
	return "1.0.0"
}

// incrementVersion increments the patch version (e.g., 1.0.0 -> 1.0.1)
func incrementVersion(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) < 3 {
		return "1.0.1"
	}
	
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return "1.0.1"
	}
	
	parts[2] = strconv.Itoa(patch + 1)
	return strings.Join(parts, ".")
}

// GetVersionInfo returns version information as a map
func GetVersionInfo() map[string]string {
	return map[string]string{
		"version":    Version,
		"build_time": BuildTime,
		"git_commit": GitCommit,
		"build_env":  BuildEnv,
		"go_version": fmt.Sprintf("go1.23"),
	}
}

// GetVersionString returns a formatted version string
func GetVersionString() string {
	return fmt.Sprintf("%s-%s (%s) built at %s", Version, BuildEnv, GitCommit, BuildTime)
}
