# spank

Slap your MacBook, it yells back.

Uses the Apple Silicon accelerometer (Bosch BMI286 IMU via IOKit HID) to detect physical hits on your laptop and plays audio responses. Single binary, no dependencies.

## Requirements

- macOS on Apple Silicon (M2+)
- `sudo` (for IOKit HID accelerometer access)

## Install

Download from the [latest release](https://github.com/taigrr/spank/releases/latest).

Or build from source:

```bash
go install github.com/taigrr/spank@latest
```

## Usage

```bash
# Normal mode — says "ow!" when slapped
sudo spank

# Sexy mode — escalating responses based on slap frequency
sudo spank --sexy
```

### Modes

**Pain mode** (default): Randomly plays from 10 pain/protest audio clips when a slap is detected.

**Sexy mode** (`--sexy`): Tracks slaps within a rolling 5-minute window. The more you slap, the more intense the audio response. 60 levels of escalation.

## How it works

1. Reads raw accelerometer data directly via IOKit HID (Apple SPU sensor)
2. Runs vibration detection (STA/LTA, CUSUM, kurtosis, peak/MAD)
3. When a significant impact is detected, plays an embedded MP3 response
4. 500ms cooldown between responses to prevent rapid-fire

## What people are saying

> "this is the most amazing thing i've ever seen" — [@kenwheeler](https://x.com/kenwheeler)

> "I just ran sexy mode with my wife sitting next to me...We died laughing" — [@duncanthedev](https://x.com/duncanthedev)

> "peak engineering" — [@tylertaewook](https://x.com/tylertaewook)

## Credits

Sensor reading and vibration detection ported from [olvvier/apple-silicon-accelerometer](https://github.com/olvvier/apple-silicon-accelerometer).

## License

MIT
