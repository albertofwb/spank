//go:build linux
// +build linux

package main

import (
	"fmt"
	"time"

	"github.com/gordonklaus/portaudio"
)

const (
	sampleRate      = 16000
	channels        = 1
	framesPerBuffer = 1024
	defaultThreshold = 2000
)

func init() {
	newPlatformSensor = newLinuxSensor
}

// linuxSensor uses microphone input to detect "slaps" (loud sounds)
type linuxSensor struct {
	threshold      int
	cooldown       time.Duration
	lastEvent      time.Time
	stream         *portaudio.Stream
	buffer         []int16
	eventActive    bool       // true when sound is currently above threshold (event in progress)
	eventEndTime   time.Time  // when the current event ended (for debounce)
}

// newLinuxSensor creates a Linux sensor with configurable threshold
func newLinuxSensor(threshold int) (platformSensor, error) {
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

	sensor := &linuxSensor{
		threshold:    thr,
		cooldown:     500 * time.Millisecond, // Cooldown after event ends
		buffer:       make([]int16, framesPerBuffer),
		eventEndTime: time.Now().Add(-time.Hour), // Initialize to past time so first detection works
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

func (s *linuxSensor) Read() (eventDetected bool, severity string, amplitude float64, err error) {
	// Skip detection if we're currently playing audio (avoid self-triggering)
	if isAudioPlaying() {
		// Still read from stream to keep buffer fresh, but ignore the data
		if err := s.stream.Read(); err != nil {
			return false, "", 0, fmt.Errorf("audio read error: %w", err)
		}
		// If we were in an event, mark it as ended
		if s.eventActive {
			s.eventActive = false
			s.eventEndTime = time.Now()
		}
		return false, "MUTED", 0, nil
	}

	// Check cooldown (time since last event ended)
	if time.Since(s.eventEndTime) < s.cooldown {
		// Still need to read to keep buffer fresh
		if err := s.stream.Read(); err != nil {
			return false, "", 0, fmt.Errorf("audio read error: %w", err)
		}
		return false, "", 0, nil
	}

	// Read audio from stream
	if err := s.stream.Read(); err != nil {
		return false, "", 0, fmt.Errorf("audio read error: %w", err)
	}

	// Calculate peak amplitude
	peakAmp := s.calculatePeakAmplitude()
	now := time.Now()

	if peakAmp > s.threshold {
		// Sound is above threshold
		if s.eventActive {
			// Event already in progress, don't trigger again
			return false, "", 0, nil
		}
		// New event started
		s.eventActive = true
		s.lastEvent = now
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

	// Sound is below threshold
	if s.eventActive {
		// Event just ended
		s.eventActive = false
		s.eventEndTime = now
	}

	return false, "", 0, nil
}

func (s *linuxSensor) calculatePeakAmplitude() int {
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

func (s *linuxSensor) Close() error {
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
		return fmt.Errorf("PortAudio not available (install libportaudio2-dev portaudio19-dev): %w", err)
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

	// Don't terminate here - let newLinuxSensor handle it
	portaudio.Terminate()
	return nil
}

// setThreshold allows adjusting the detection threshold
func (s *linuxSensor) setThreshold(threshold int) {
	s.threshold = threshold
}

// getDeviceInfo returns information about available audio devices
func getDeviceInfo() ([]string, error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, err
	}
	defer portaudio.Terminate()

	devices, err := portaudio.Devices()
	if err != nil {
		return nil, err
	}

	var info []string
	for i, dev := range devices {
		if dev.MaxInputChannels > 0 {
			info = append(info, fmt.Sprintf("[%d] %s (channels: %d, rate: %f)",
				i, dev.Name, dev.MaxInputChannels, dev.DefaultSampleRate))
		}
	}

	return info, nil
}
