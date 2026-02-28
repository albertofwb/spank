# spank

[简体中文](README_CN.md) | English

Slap your MacBook, it yells back.

> "this is the most amazing thing i've ever seen" — [@kenwheeler](https://x.com/kenwheeler)

> "I just ran sexy mode with my wife sitting next to me...We died laughing" — [@duncanthedev](https://x.com/duncanthedev)

> "peak engineering" — [@tylertaewook](https://x.com/tylertaewook)

Uses sensors to detect physical hits on your laptop and plays audio responses. Single binary, cross-platform.

## Requirements

### macOS
- macOS on Apple Silicon (M2+)
- `sudo` (for IOKit HID accelerometer access)

### Linux
- Linux with PortAudio support
- `libportaudio2-dev` and `portaudio19-dev` packages
- Microphone (to detect slaps via audio)

### Windows
- Windows 10/11
- Microphone (to detect slaps via audio)
- No additional dependencies required

## Install

Download from the [latest release](https://github.com/albertofwb/spank/releases/latest).

Or build from source:

```bash
go install github.com/taigrr/spank@latest
```

Or clone and build with Make:

```bash
git clone https://github.com/albertofwb/spank.git
cd spank
make build
```

## Usage

### macOS

```bash
# Normal mode — says "ow!" when slapped
sudo spank

# Sexy mode — escalating responses based on slap frequency
sudo spank --sexy

# Halo mode — plays Halo death sounds when slapped
sudo spank --halo
```

### Linux

```bash
# Normal mode — says "ow!" when you make a loud sound
spank

# Sexy mode — escalating responses
spank --sexy

# Halo mode — Halo death sounds
spank --halo

# Adjust detection threshold (default: 2000, higher = less sensitive)
spank --threshold 3000
```

> **Linux Note:** The microphone is used to detect loud sounds (like slapping the laptop). Make sure your microphone is not muted and has reasonable volume.
>
> **Hide ALSA warnings:** PortAudio may output ALSA debug messages. Run with `2>/dev/null` to hide them:
> ```bash
> spank 2>/dev/null
> spank --threshold 3000 2>/dev/null
> ```

### Windows

```powershell
# Normal mode — says "ow!" when you make a loud sound
spank.exe

# Sexy mode — escalating responses
spank.exe --sexy

# Halo mode — Halo death sounds
spank.exe --halo

# Adjust detection threshold (default: 2000, higher = less sensitive)
spank.exe --threshold 3000
```

> **Windows Note:** The microphone is used to detect loud sounds. Make sure your microphone is enabled in Windows Privacy settings.
>
> **Threshold tuning:** If detection is too sensitive, increase `--threshold` (e.g., 3000-5000). If not sensitive enough, decrease it (e.g., 1000-1500).
>
> **Hide ALSA warnings:** PortAudio may output ALSA debug messages. Run with `2>/dev/null` to hide them:
> ```bash
> spank 2>/dev/null
> spank --threshold 3000 2>/dev/null
> ```

### Modes

**Pain mode** (default): Randomly plays from 10 pain/protest audio clips when a slap is detected.

**Sexy mode** (`--sexy`): Tracks slaps within a rolling 5-minute window. The more you slap, the more intense the audio response. 60 levels of escalation.

**Halo mode** (`--halo`): Randomly plays from death sound effects from the Halo video game series when a slap is detected.

## Building

### Auto-detect platform

```bash
make build
```

### Cross-compile (from Linux)

```bash
make cross-compile
```

### Install to system

```bash
make install
```

### Run directly

```bash
make run          # Normal mode
make run-sexy     # Sexy mode
make run-halo     # Halo mode
```

## Running as a Service (macOS)

To have spank start automatically at boot, create a launchd plist. Pick your mode:

<details>
<summary>Pain mode (default)</summary>

```bash
sudo tee /Library/LaunchDaemons/com.taigrr.spank.plist > /dev/null << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.taigrr.spank</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/spank</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/spank.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/spank.err</string>
</dict>
</plist>
EOF
```

</details>

<details>
<summary>Sexy mode</summary>

```bash
sudo tee /Library/LaunchDaemons/com.taigrr.spank.plist > /dev/null << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.taigrr.spank</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/spank</string>
        <string>--sexy</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/spank.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/spank.err</string>
</dict>
</plist>
EOF
```

</details>

<details>
<summary>Halo mode</summary>

```bash
sudo tee /Library/LaunchDaemons/com.taigrr.spank.plist > /dev/null << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.taigrr.spank</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/spank</string>
        <string>--halo</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/spank.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/spank.err</string>
</dict>
</plist>
EOF
```

</details>

> **Note:** Update the path to `spank` if you installed it elsewhere (e.g. `~/go/bin/spank`).

Load and start the service:

```bash
sudo launchctl load /Library/LaunchDaemons/com.taigrr.spank.plist
```

Since the plist lives in `/Library/LaunchDaemons` and no `UserName` key is set, launchd runs it as root — no `sudo` needed.

To stop or unload:

```bash
sudo launchctl unload /Library/LaunchDaemons/com.taigrr.spank.plist
```

## How it works

### macOS
1. Reads raw accelerometer data directly via IOKit HID (Apple SPU sensor - Bosch BMI286 IMU)
2. Runs vibration detection (STA/LTA, CUSUM, kurtosis, peak/MAD)
3. When a significant impact is detected, plays an embedded MP3 response
4. 500ms cooldown between responses to prevent rapid-fire

### Linux
1. Captures audio from microphone using PortAudio (continuous stream)
2. Analyzes audio amplitude in real-time to detect loud sounds (slaps)
3. When a sound exceeds the threshold, plays an embedded MP3 response
4. 500ms cooldown between responses

## Platform Differences

| Feature | macOS | Linux | Windows |
|---------|-------|-------|---------|
| Sensor | Accelerometer (IOKit HID) | Microphone (PortAudio) | Microphone (PortAudio) |
| Requires sudo | Yes (IOKit access) | No | No |
| Hardware | Apple Silicon M2+ | Any with microphone | Any with microphone |
| Trigger | Physical impact | Loud sound | Loud sound |

## Credits

- Sensor reading and vibration detection ported from [olvvier/apple-silicon-accelerometer](https://github.com/olvvier/apple-silicon-accelerometer)
- Cross-platform adaptation with microphone support for Linux

## License

MIT
