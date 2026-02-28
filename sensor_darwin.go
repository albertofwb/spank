//go:build darwin
// +build darwin

package main

import (
	"fmt"
	"time"

	"github.com/taigrr/apple-silicon-accelerometer/detector"
	"github.com/taigrr/apple-silicon-accelerometer/sensor"
	"github.com/taigrr/apple-silicon-accelerometer/shm"
)

// sensorReady is closed once shared memory is created and the sensor
// worker is about to enter the CFRunLoop.
var sensorReady = make(chan struct{})

// sensorErr receives any error from the sensor worker.
var sensorErr = make(chan error, 1)

func init() {
	newPlatformSensor = newMacOSSensor
}

type macOSSensor struct {
	accelRing    *shm.Ring
	lastAccelTotal uint64
	det          *detector.Detector
}

// newMacOSSensor creates a macOS sensor (threshold ignored on macOS)
func newMacOSSensor(threshold int) (platformSensor, error) {
	// Create shared memory for accelerometer data.
	accelRing, err := shm.CreateRing(shm.NameAccel)
	if err != nil {
		return nil, fmt.Errorf("creating accel shm: %w", err)
	}

	// Start the sensor worker in a background goroutine.
	go func() {
		close(sensorReady)
		err := sensor.Run(sensor.Config{
			AccelRing: accelRing,
			Restarts:  0,
		})
		if err != nil {
			sensorErr <- err
		}
	}()

	// Wait for sensor to be ready.
	select {
	case <-sensorReady:
	case err := <-sensorErr:
		accelRing.Close()
		accelRing.Unlink()
		return nil, fmt.Errorf("sensor worker failed: %w", err)
	case <-time.After(5 * time.Second):
		accelRing.Close()
		accelRing.Unlink()
		return nil, fmt.Errorf("sensor timeout")
	}

	// Give the sensor a moment to start producing data.
	time.Sleep(100 * time.Millisecond)

	return &macOSSensor{
		accelRing:    accelRing,
		det:          detector.New(),
	}, nil
}

func (s *macOSSensor) Read() (eventDetected bool, severity string, amplitude float64, err error) {
	select {
	case err := <-sensorErr:
		return false, "", 0, fmt.Errorf("sensor worker failed: %w", err)
	default:
	}

	tNow := float64(time.Now().UnixNano()) / 1e9

	samples, newTotal := s.accelRing.ReadNew(s.lastAccelTotal, shm.AccelScale)
	s.lastAccelTotal = newTotal

	maxBatch := 200
	if len(samples) > maxBatch {
		samples = samples[len(samples)-maxBatch:]
	}

	nSamples := len(samples)
	for idx, sample := range samples {
		tSample := tNow - float64(nSamples-idx-1)/float64(s.det.FS)
		s.det.Process(sample.X, sample.Y, sample.Z, tSample)
	}

	newEventIdx := len(s.det.Events)
	if newEventIdx > 0 {
		ev := s.det.Events[newEventIdx-1]
		if ev.Severity == "CHOC_MAJEUR" || ev.Severity == "CHOC_MOYEN" || ev.Severity == "MICRO_CHOC" {
			return true, ev.Severity, ev.Amplitude, nil
		}
	}

	return false, "", 0, nil
}

func (s *macOSSensor) Close() error {
	if s.accelRing != nil {
		s.accelRing.Close()
		s.accelRing.Unlink()
	}
	return nil
}

func checkPlatformRequirements() error {
	// macOS requires root for IOKit HID access
	return nil
}
