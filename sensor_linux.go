//go:build linux
// +build linux

package main

func init() {
	newPlatformSensor = newLinuxSensor
}

// newLinuxSensor creates a Linux microphone sensor
func newLinuxSensor(threshold int) (platformSensor, error) {
	return newMicrophoneSensor(threshold)
}
