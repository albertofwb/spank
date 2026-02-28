// spank detects slaps/hits on the laptop and plays audio responses.
// It reads the Apple Silicon accelerometer directly via IOKit HID —
// no separate sensor daemon required. Needs sudo.
//
// Cross-platform support:
//   - macOS: Uses Apple Silicon accelerometer (IOKit HID)
//   - Linux: Uses microphone to detect loud sounds (slaps)
//go:build darwin || linux
// +build darwin linux

package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/charmbracelet/fang"
	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/spf13/cobra"
)

var version = "dev"

//go:embed audio/pain/*.mp3
var painAudio embed.FS

//go:embed audio/sexy/*.mp3
var sexyAudio embed.FS

//go:embed audio/halo/*.mp3
var haloAudio embed.FS

var (
	sexyMode  bool
	haloMode  bool
	threshold int  // Linux only: microphone threshold
)

// platformSensor is the interface for platform-specific sensor implementations
type platformSensor interface {
	// Read returns (eventDetected, severity, amplitude, error)
	Read() (bool, string, float64, error)
	Close() error
}

// newPlatformSensor creates a platform-specific sensor with optional threshold (Linux only)
var newPlatformSensor func(threshold int) (platformSensor, error)

type playMode int

const (
	modeRandom playMode = iota
	modeEscalation
)

type soundPack struct {
	name  string
	fs    embed.FS
	dir   string
	mode  playMode
	files []string
}

func (sp *soundPack) loadFiles() error {
	entries, err := sp.fs.ReadDir(sp.dir)
	if err != nil {
		return err
	}
	sp.files = make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			sp.files = append(sp.files, sp.dir+"/"+e.Name())
		}
	}
	sort.Strings(sp.files)
	return nil
}

type slapTracker struct {
	mu     sync.Mutex
	times  []time.Time
	window time.Duration
	pack   *soundPack
	altIdx int
}

func newSlapTracker(pack *soundPack) *slapTracker {
	return &slapTracker{
		window: 5 * time.Minute,
		pack:   pack,
	}
}

func (st *slapTracker) record(t time.Time) int {
	st.mu.Lock()
	defer st.mu.Unlock()

	cutoff := t.Add(-st.window)
	newTimes := make([]time.Time, 0, len(st.times)+1)
	for _, tt := range st.times {
		if tt.After(cutoff) {
			newTimes = append(newTimes, tt)
		}
	}
	newTimes = append(newTimes, t)
	st.times = newTimes
	return len(st.times)
}

func (st *slapTracker) getFile(count int) string {
	st.mu.Lock()
	defer st.mu.Unlock()

	if len(st.pack.files) == 0 {
		return ""
	}

	if st.pack.mode == modeRandom {
		return st.pack.files[rand.Intn(len(st.pack.files))]
	}

	// Escalation mode
	maxIdx := len(st.pack.files) - 1
	topTwo := maxIdx - 1
	if topTwo < 0 {
		topTwo = 0
	}

	var idx int
	if count >= 20 {
		st.altIdx = 1 - st.altIdx
		idx = topTwo + st.altIdx
	} else {
		ratio := float64(count) / 20.0
		if ratio > 1 {
			ratio = 1
		}
		idx = int(ratio * float64(topTwo))
	}

	if idx > maxIdx {
		idx = maxIdx
	}
	return st.pack.files[idx]
}

func main() {
	cmd := &cobra.Command{
		Use:   "spank",
		Short: "Yells 'ow!' when you slap the laptop",
		Long: `spank detects slaps/hits on the laptop and plays audio responses.

Platform-specific behavior:
  macOS: Uses Apple Silicon accelerometer (IOKit HID) - requires root
  Linux: Uses microphone to detect loud sounds (slaps) - no root needed

Use --sexy for a different experience. In sexy mode, the more you slap
within a minute, the more intense the sounds become.

Use --halo to play random audio clips from Halo soundtracks on each slap.`,
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context())
		},
		SilenceUsage: true,
	}

	cmd.Flags().BoolVarP(&sexyMode, "sexy", "s", false, "Enable sexy mode")
	cmd.Flags().BoolVarP(&haloMode, "halo", "H", false, "Enable halo mode")
	cmd.Flags().IntVar(&threshold, "threshold", 0, "Microphone detection threshold (Linux only, default: 2000)")

	if err := fang.Execute(context.Background(), cmd); err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Platform-specific requirements check
	if err := checkPlatformRequirements(); err != nil {
		return err
	}

	if sexyMode && haloMode {
		return fmt.Errorf("--sexy and --halo are mutually exclusive; pick one")
	}

	var pack *soundPack
	switch {
	case sexyMode:
		pack = &soundPack{name: "sexy", fs: sexyAudio, dir: "audio/sexy", mode: modeEscalation}
	case haloMode:
		pack = &soundPack{name: "halo", fs: haloAudio, dir: "audio/halo", mode: modeRandom}
	default:
		pack = &soundPack{name: "pain", fs: painAudio, dir: "audio/pain", mode: modeRandom}
	}

	if err := pack.loadFiles(); err != nil {
		return fmt.Errorf("loading %s audio: %w", pack.name, err)
	}

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Create platform-specific sensor
	sensor, err := newPlatformSensor(threshold)
	if err != nil {
		return fmt.Errorf("initializing sensor: %w", err)
	}
	defer sensor.Close()

	tracker := newSlapTracker(pack)
	speakerInit := false
	lastYell := time.Time{}
	cooldown := 500 * time.Millisecond

	fmt.Printf("spank: listening for slaps in %s mode... (ctrl+c to quit)\n", pack.name)

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nbye!")
			return nil
		case <-ticker.C:
		}

		now := time.Now()

		// Platform-specific sensor read
		eventDetected, severity, amplitude, err := sensor.Read()
		if err != nil {
			return fmt.Errorf("sensor error: %w", err)
		}

		if eventDetected {
			if time.Since(lastYell) > cooldown {
				lastYell = now
				count := tracker.record(now)
				file := tracker.getFile(count)
				fmt.Printf("slap #%d [%s amp=%.5fg] -> %s\n", count, severity, amplitude, file)
				go playEmbedded(pack.fs, file, &speakerInit)
			}
		}
	}
}

var speakerMu sync.Mutex

func playEmbedded(fs embed.FS, path string, speakerInit *bool) {
	data, err := fs.ReadFile(path)
	if err != nil {
		return
	}

	streamer, format, err := mp3.Decode(io.NopCloser(bytes.NewReader(data)))
	if err != nil {
		return
	}
	defer streamer.Close()

	speakerMu.Lock()
	if !*speakerInit {
		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		*speakerInit = true
	}
	speakerMu.Unlock()

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))
	<-done
}
