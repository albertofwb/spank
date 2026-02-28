//go:build windows
// +build windows

package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed portaudio.dll
var portaudioDLL []byte

func init() {
	// Extract DLL to temp directory on startup
	if err := extractDLL(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to extract portaudio.dll: %v\n", err)
	}
	newPlatformSensor = newWindowsSensor
}

// extractDLL extracts the embedded portaudio.dll to a temp directory
// and adds that directory to PATH so Windows can find it
func extractDLL() error {
	// Create temp directory for DLL
	tempDir := filepath.Join(os.TempDir(), "spank")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}

	// Write DLL if not exists or update it
	dllPath := filepath.Join(tempDir, "portaudio.dll")
	if err := os.WriteFile(dllPath, portaudioDLL, 0644); err != nil {
		return fmt.Errorf("write dll: %w", err)
	}

	// Add temp directory to PATH so Windows can find the DLL
	currentPath := os.Getenv("PATH")
	os.Setenv("PATH", tempDir+";"+currentPath)

	return nil
}

// newWindowsSensor creates a Windows microphone sensor
func newWindowsSensor(threshold int) (platformSensor, error) {
	return newMicrophoneSensor(threshold)
}
