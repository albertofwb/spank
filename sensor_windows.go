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
	// Extract DLL to exe directory on startup (before PortAudio initializes)
	if err := extractDLL(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to extract portaudio.dll: %v\n", err)
	}
	newPlatformSensor = newWindowsSensor
}

// extractDLL extracts the embedded portaudio.dll to the same directory as the exe
// This ensures Windows can find it when loading PortAudio
func extractDLL() error {
	// Get the directory of the current executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}
	exeDir := filepath.Dir(exePath)

	// Write DLL to the same directory as the executable
	dllPath := filepath.Join(exeDir, "portaudio.dll")

	// Check if DLL already exists with same size (avoid rewriting)
	if info, err := os.Stat(dllPath); err == nil && info.Size() == int64(len(portaudioDLL)) {
		// DLL already exists with correct size, skip extraction
		return nil
	}

	if err := os.WriteFile(dllPath, portaudioDLL, 0644); err != nil {
		return fmt.Errorf("write dll to %s: %w", dllPath, err)
	}

	return nil
}

// newWindowsSensor creates a Windows microphone sensor
func newWindowsSensor(threshold int) (platformSensor, error) {
	return newMicrophoneSensor(threshold)
}
