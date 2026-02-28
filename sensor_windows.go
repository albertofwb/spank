//go:build windows
// +build windows

package main

import (
	"fmt"
	"time"

	"github.com/gordonklaus/portaudio"
)

const (
	sampleRate       = 16000
	channels         = 1
	framesPerBuffer  = 1024
	defaultThreshold = 2000
)

func init() {
	newPlatformSensor = newWindowsSensor
}

// windowsSensor uses microphone input to detect "slaps" (loud sounds)
type windowsSensor struct {
	threshold int
	cooldown  time.Duration
	lastEvent time.Time
	stream    *portaudio.Stream
	buffer    []int16
}

// newWindowsSensor creates a Windows sensor with configurable threshold
func newWindowsSensor(threshold int) (platformSensor, error) {
	// Check platform requirements first
	if err := checkPlatformRequirements(); err != nil {
		return nil, err
	}

	// Initialize PortAudio
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize PortAudio: %w", err)
	}

	// Use provided threshold or default
	thr := threshold
	if thr <= 0 {
		thr = defaultThreshold
	}

	sensor := &windowsSensor{
		threshold: thr,
		cooldown:  500 * time.Millisecond,
		buffer:    make([]int16, framesPerBuffer),
	}

	// Open default input stream
	stream, err := portaudio.OpenDefaultStream(
		channels,      // input channels
		0,             // output channels
		sampleRate,    // sample rate
		framesPerBuffer,
		sensor.buffer, // buffer to fill
	)
	if err != nil {
		portaudio.Terminate()
		return nil, fmt.Errorf("failed to open audio stream: %w", err)
	}

	sensor.stream = stream

	// Start the stream
	if err := stream.Start(); err != nil {
		stream.Close()
		portaudio.Terminate()
		return nil, fmt.Errorf("failed to start audio stream: %w", err)
	}

	return sensor, nil
}

func (s *windowsSensor) Read() (eventDetected bool, severity string, amplitude float64, err error) {
	// Check cooldown
	if time.Since(s.lastEvent) < s.cooldown {
		return false, "", 0, nil
	}

	// Read audio from stream
	if err := s.stream.Read(); err != nil {
		return false, "", 0, fmt.Errorf("audio read error: %w", err)
	}

	// Calculate peak amplitude
	peakAmp := s.calculatePeakAmplitude()

	if peakAmp > s.threshold {
		s.lastEvent = time.Now()
		// Normalize amplitude to a scale similar to accelerometer g-force
		amp := float64(peakAmp) / 32768.0 * 3.0

		// Determine severity based on amplitude
		sev := "MICRO_CHOC"
		if amp > 1.5 {
			sev = "CHOC_MAJEUR"
		} else if amp > 0.8 {
			sev = "CHOC_MOYEN"
		}

		return true, sev, amp, nil
	}

	return false, "", 0, nil
}

func (s *windowsSensor) calculatePeakAmplitude() int {
	maxAmp := 0
	for _, sample := range s.buffer {
		val := int(sample)
		if val < 0 {
			val = -val
		}
		if val > maxAmp {
			maxAmp = val
		}
	}
	return maxAmp
}

func (s *windowsSensor) Close() error {
	if s.stream != nil {
		s.stream.Stop()
		s.stream.Close()
	}
	portaudio.Terminate()
	return nil
}

func checkPlatformRequirements() error {
	// Check if PortAudio is available
	if err := portaudio.Initialize(); err != nil {
		return fmt.Errorf("PortAudio not available: %w", err)
	}

	// Try to list devices to verify microphone exists
	devices, err := portaudio.Devices()
	if err != nil {
		portaudio.Terminate()
		return fmt.Errorf("failed to list audio devices: %w", err)
	}

	hasInput := false
	for _, dev := range devices {
		if dev.MaxInputChannels > 0 {
			hasInput = true
			break
		}
	}

	if !hasInput {
		portaudio.Terminate()
		return fmt.Errorf("no audio input devices found. Check your microphone connection.")
	}

	portaudio.Terminate()
	return nil
}
