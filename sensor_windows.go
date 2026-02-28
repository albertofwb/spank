//go:build windows
// +build windows

package main

func init() {
	newPlatformSensor = newWindowsSensor
}

// newWindowsSensor creates a Windows microphone sensor
func newWindowsSensor(threshold int) (platformSensor, error) {
	return newMicrophoneSensor(threshold)
}
