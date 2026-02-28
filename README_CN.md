# spank

[English](README.md) | 简体中文

拍你的 MacBook，它会叫出声。

> "这是我见过最神奇的东西" — [@kenwheeler](https://x.com/kenwheeler)

> "我刚开始运行 sexy 模式，我老婆坐在我旁边...我们都笑死了" — [@duncanthedev](https://x.com/duncanthedev)

> "巅峰工程" — [@tylertaewook](https://x.com/tylertaewook)

使用传感器检测笔记本电脑上的物理撞击并播放音频回应。单一二进制文件，跨平台支持。

## 系统要求

### macOS
- Apple Silicon 芯片的 macOS (M2+)
- `sudo` 权限（用于 IOKit HID 加速度计访问）

### Linux
- 支持 PortAudio 的 Linux 系统
- `libportaudio2-dev` 和 `portaudio19-dev` 包
- 麦克风（用于通过音频检测拍打）

## 安装

从 [最新发布](https://github.com/taigrr/spank/releases/latest) 下载。

或者从源码构建：

```bash
go install github.com/taigrr/spank@latest
```

或者克隆并使用 Make 构建：

```bash
git clone https://github.com/taigrr/spank.git
cd spank
make build
```

## 使用方法

### macOS

```bash
# 普通模式 — 被拍打时喊 "ow!"
sudo spank

# Sexy 模式 — 根据拍打频率递进回应
sudo spank --sexy

# Halo 模式 — 播放 Halo 死亡音效
sudo spank --halo
```

### Linux

```bash
# 普通模式 — 检测到大声时喊 "ow!"
spank

# Sexy 模式 — 递进回应
spank --sexy

# Halo 模式 — Halo 死亡音效
spank --halo

# 调整检测阈值（默认：2000，越高越不敏感）
spank --threshold 3000
```

> **Linux 注意：** 程序使用麦克风检测大声响（比如拍打笔记本）。确保麦克风未静音且音量合适。
>
> **阈值调节：** 如果检测太敏感，增加 `--threshold`（比如 3000-5000）。如果不够敏感，降低它（比如 1000-1500）。
>
> **隐藏 ALSA 警告：** PortAudio 可能会输出 ALSA 调试信息。使用 `2>/dev/null` 隐藏：
> ```bash
> spank 2>/dev/null
> spank --threshold 3000 2>/dev/null
> ```

### 模式

**Pain 模式**（默认）：检测到拍打时随机播放 10 个疼痛/抗议音频片段之一。

**Sexy 模式** (`--sexy`)：在 5 分钟滚动窗口内跟踪拍打次数。拍打得越多，音频回应越强烈。60 个递进级别。

**Halo 模式** (`--halo`)：检测到拍打时随机播放 Halo 游戏系列的死亡音效。

## 构建

### 自动检测平台

```bash
make build
```

### 交叉编译（从 Linux）

```bash
make cross-compile
```

### 安装到系统

```bash
make install
```

### 直接运行

```bash
make run          # 普通模式
make run-sexy     # Sexy 模式
make run-halo     # Halo 模式
```

## 作为服务运行（macOS）

要让 spank 开机自动启动，创建 launchd plist。选择你的模式：

<details>
<summary>Pain 模式（默认）</summary>

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
<summary>Sexy 模式</summary>

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
<summary>Halo 模式</summary>

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

> **注意：** 如果你将 spank 安装在其他位置（例如 `~/go/bin/spank`），请更新路径。

加载并启动服务：

```bash
sudo launchctl load /Library/LaunchDaemons/com.taigrr.spank.plist
```

由于 plist 位于 `/Library/LaunchDaemons` 且未设置 `UserName` 键，launchd 以 root 身份运行它 —— 不需要 `sudo`。

停止或卸载：

```bash
sudo launchctl unload /Library/LaunchDaemons/com.taigrr.spank.plist
```

## 工作原理

### macOS
1. 直接通过 IOKit HID 读取原始加速度计数据（Apple SPU 传感器 - Bosch BMI286 IMU）
2. 运行振动检测（STA/LTA、CUSUM、峰度、峰值/MAD）
3. 检测到显著撞击时，播放嵌入的 MP3 回应
4. 两次回应之间有 500ms 冷却时间

### Linux
1. 使用 PortAudio 从麦克风捕获音频（连续流）
2. 实时分析音频幅度以检测大声响（拍打）
3. 当声音超过阈值时，播放嵌入的 MP3 回应
4. 两次回应之间有 500ms 冷却时间

## 平台差异

| 特性 | macOS | Linux |
|---------|-------|-------|
| 传感器 | 加速度计 (IOKit HID) | 麦克风 (PortAudio) |
| 需要 sudo | 是 (IOKit 访问) | 否 |
| 硬件 | Apple Silicon M2+ | 任何带麦克风的设备 |
| 触发方式 | 物理撞击 | 大声响 |

## 致谢

- 传感器读取和振动检测移植自 [olvvier/apple-silicon-accelerometer](https://github.com/olvvier/apple-silicon-accelerometer)
- Linux 跨平台适配，支持麦克风

## 许可证

MIT

---

[English README](README.md)
