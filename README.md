# 🎤 Tuneminal - Terminal Karaoke Machine

 ████████╗██╗   ██╗███╗   ██╗███████╗███╗   ███╗██╗███╗   ██╗ █████╗ ██╗     
 ╚══██╔══╝██║   ██║████╗  ██║██╔════╝████╗ ████║██║████╗  ██║██╔══██╗██║     
    ██║   ██║   ██║██╔██╗ ██║█████╗  ██╔████╔██║██║██╔██╗ ██║███████║██║     
    ██║   ██║   ██║██║╚██╗██║██╔══╝  ██║╚██╔╝██║██║██║╚██╗██║██╔══██║██║     
    ██║   ╚██████╔╝██║ ╚████║███████╗██║ ╚═╝ ██║██║██║ ╚████║██║  ██║███████╗
    ╚═╝    ╚═════╝ ╚═╝  ╚═══╝╚══════╝╚═╝     ╚═╝╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝╚══════╝

**Tuneminal** is a powerful command-line karaoke machine with live audio visualization built in Go. Experience a rich, animated terminal interface with synchronized lyrics, real-time audio visualization, and karaoke scoring - all from your command line!

![Go Version](https://img.shields.io/badge/Go-1.21+-blue)
![License](https://img.shields.io/badge/License-MIT-green)
![Platform](https://img.shields.io/badge/Platform-Cross--Platform-lightgrey)
![Downloads](https://img.shields.io/github/downloads/tuneminal/tuneminal/total)
![Stars](https://img.shields.io/github/stars/tuneminal/tuneminal)

## ✨ Features

- 🎵 **Rich Terminal Interface**: Beautiful animated UI with real-time updates
- 🎤 **Karaoke Mode**: Synchronized lyrics with scoring system and streak tracking
- 🎨 **Live Audio Visualizer**: Real-time spectrum analysis with dynamic bars
- 📝 **LRC Lyrics Support**: Time-coded lyrics with automatic scrolling
- 🔍 **Smart Search**: Filter songs by title or artist with real-time results
- 🎵 **Audio Playback**: MP3 and WAV support with play/pause/stop controls
- 📊 **Performance Scoring**: Earn points for singing accuracy and build streaks
- 🖥️ **Responsive Layout**: Adapts to terminal size with professional styling
- ⌨️ **Intuitive Controls**: Easy keyboard navigation and shortcuts
- 🚀 **Cross-Platform**: Works on Windows, macOS, and Linux

## 📸 Screenshots

### Main Interface
![Main Interface](assets/main%20screen.png)

### Loading Screen
![Loading Screen](assets/loading%20screen.png)

### Help Window
![Help Window](assets/help%20window.png)

### Live Demo
![Live Demo](assets/playback.gif)

## 🚀 Installation

> **Note**: This is a local development project. The installation methods below assume you have the source code locally.

### Quick Install (Recommended)

#### Windows
```powershell
# Option 1: Use the local installer script
.\install.ps1

# Option 2: Build from source
.\build.ps1 build
.\tuneminal.exe
```

#### macOS
```bash
# Build from source (recommended)
git clone <your-repo-url> tuneminal
cd tuneminal
go mod tidy
go build -o tuneminal cmd/tuneminal/main.go
./tuneminal
```

#### Linux
```bash
# Build from source (recommended)
git clone <your-repo-url> tuneminal
cd tuneminal
go mod tidy
go build -o tuneminal cmd/tuneminal/main.go
./tuneminal
```

### Build from Source

#### Prerequisites
- **Go 1.21+**: Download from [golang.org](https://golang.org/dl/)
- **Audio libraries**: ALSA (Linux), Core Audio (macOS), DirectSound (Windows)

#### Build Steps

1. **Clone the repository**:
   ```bash
   git clone https://github.com/tuneminal/tuneminal.git
   cd tuneminal
   ```

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

3. **Build the application**:
   ```bash
   # Development build
   go build -o tuneminal cmd/tuneminal/main.go
   
   # Optimized release build
   go build -ldflags="-s -w" -o tuneminal cmd/tuneminal/main.go
   ```

4. **Run Tuneminal**:
   ```bash
   ./tuneminal
   ```

### Docker Installation

```bash
# Build Docker image locally
docker build -t tuneminal .

# Run with audio support (Linux)
docker run -it --device /dev/snd tuneminal

# Run without audio (for testing)
docker run -it tuneminal
```

## 🎮 Usage

### Getting Started

1. **Launch Tuneminal**:
   ```bash
   tuneminal
   ```

2. **Add Your Music**: Place MP3/WAV files in `uploads/demo/` directory
3. **Select a Song**: Use arrow keys to navigate and Enter to play
4. **Sing Along**: Watch the synchronized lyrics and earn points!

### Interface Overview

The application features several key sections:

- **🎵 Music Library**: Browse and search your song collection
- **🎤 Now Playing**: Current song information and playback status
- **🎨 Audio Visualizer**: Real-time spectrum analysis
- **📝 Karaoke Lyrics**: 5-line centered lyric display with scoring
- **📊 Score Panel**: Performance tracking and streak counter

### Keyboard Controls

#### Navigation
- `↑/↓`: Navigate song list
- `Tab`: Switch between search and song list
- `Enter`: Play selected song
- `Q`: Quit application

#### Search
- `/`: Focus search box
- `Esc`: Clear search and return to song list
- Type to filter songs by title or artist

#### Playback
- `Space`: Play/Pause current song
- `S`: Stop playback
- `N`: Next song
- `P`: Previous song

#### Help
- `H`: Show/hide help window
- `L`: Focus lyrics panel
- `R`: Reload song library

### File Organization

Place your audio and lyrics files in the `uploads/demo/` directory:

```
uploads/
└── demo/
    ├── song1.mp3          # Audio file
    ├── song1.lrc          # Lyrics file (same name)
    ├── song2.wav          # Another audio file
    └── song2.lrc          # Corresponding lyrics
```

### LRC Lyrics Format

Create `.lrc` files with time-coded lyrics for synchronized karaoke:

```lrc
[ar:Artist Name]
[ti:Song Title]
[al:Album Name]

[00:00.50]First line of lyrics
[00:03.25]Second line of lyrics
[00:06.00]Third line of lyrics
[00:08.75]Fourth line of lyrics
```

**Time Format**: `[mm:ss.xx]` where:
- `mm`: Minutes (00-99)
- `ss`: Seconds (00-59)  
- `xx`: Centiseconds (00-99, 1/100th of a second)

**Tips for LRC files**:
- Use the same filename as your audio file (e.g., `song.mp3` → `song.lrc`)
- Time codes should be in chronological order
- Empty lines `[]` create pauses in the display
- Metadata tags `[ar:]`, `[ti:]`, `[al:]` are optional but recommended

## 🏗️ Project Structure

```
tuneminal/
├── cmd/
│   └── tuneminal/
│       └── main.go          # Application entry point
├── pkg/
│   ├── player/              # Audio playback engine
│   │   ├── player.go       # Audio player implementation
│   │   ├── lyrics.go       # LRC lyrics parser
│   │   └── player_test.go  # Unit tests
│   ├── metadata/            # Audio metadata extraction
│   │   └── metadata.go     # File metadata reader
│   └── utils/              # Utility functions
│       └── files.go        # File system helpers
├── uploads/
│   └── demo/               # Demo files directory
│       ├── *.mp3           # Sample audio files
│       └── *.lrc           # Sample lyrics files
├── packaging/              # Distribution packages
│   ├── chocolatey/         # Windows package
│   ├── Formula/           # Homebrew formula
│   └── snap/              # Snap package
├── go.mod                  # Go module definition
├── Dockerfile             # Container configuration
├── install.ps1            # Windows installer
├── build.ps1              # Build script
└── README.md              # This file
```

## 🛠️ Development

### Dependencies

- **[tview](https://github.com/rivo/tview)**: Terminal UI framework
- **[tcell](https://github.com/gdamore/tcell)**: Terminal cell manipulation
- **[Beep](https://github.com/faiface/beep)**: Audio playback library
- **[Oto](https://github.com/ebitengine/oto)**: Cross-platform audio library

### Building

```bash
# Development build
go build -o tuneminal cmd/tuneminal/main.go

# Release build (optimized)
go build -ldflags="-s -w" -o tuneminal cmd/tuneminal/main.go

# Cross-platform builds
GOOS=windows GOARCH=amd64 go build -o tuneminal.exe cmd/tuneminal/main.go
GOOS=linux GOARCH=amd64 go build -o tuneminal-linux cmd/tuneminal/main.go
GOOS=darwin GOARCH=amd64 go build -o tuneminal-macos cmd/tuneminal/main.go
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/player/
```

### Development Scripts

```bash
# Windows PowerShell
.\build.ps1 build      # Build application
.\build.ps1 run        # Build and run
.\build.ps1 test       # Run tests
.\build.ps1 clean      # Clean artifacts

# Unix Makefile
make build             # Build application
make run              # Build and run
make test             # Run tests
make clean            # Clean artifacts
```

## 🎵 Demo

### Video Demonstration
Watch Tuneminal in action:

https://github.com/tuneminal/tuneminal/assets/playback.mp4

### Try the included demo files:

1. **Demo files included**: Check `uploads/demo/` directory
2. **Launch Tuneminal**: Run `tuneminal` command
3. **Select a song**: Use arrow keys to navigate and Enter to play
4. **Enjoy karaoke**: Watch synchronized lyrics and visualizer!

**Sample files included**:
- `Heroes Tonight.mp3` - Sample audio file
- `Heroes Tonight.lrc` - Synchronized lyrics

## 🐛 Troubleshooting

### Common Issues

#### Audio Not Playing
- **Check file format**: Ensure files are MP3 or WAV
- **Verify file permissions**: Make sure files are readable
- **Audio drivers**: Ensure system audio is working
- **File path**: Check files are in `uploads/demo/` directory

#### Lyrics Not Syncing
- **LRC format**: Verify time format `[mm:ss.xx]`
- **File naming**: Lyrics file must match audio filename
- **Time order**: Ensure time codes are chronological
- **File encoding**: Use UTF-8 encoding for lyrics files

#### Visualizer Not Working
- **Terminal support**: Ensure terminal supports Unicode
- **Window size**: Try resizing terminal window
- **Audio playback**: Visualizer only works during playback

#### Build Issues
- **Go version**: Ensure Go 1.21+ is installed
- **Dependencies**: Run `go mod tidy` to update dependencies
- **Audio libraries**: Install platform-specific audio libraries

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Setup

1. **Fork the repository**
2. **Clone your fork**: `git clone https://github.com/yourusername/tuneminal.git`
3. **Create a feature branch**: `git checkout -b feature/amazing-feature`
4. **Make your changes** and add tests if applicable
5. **Commit your changes**: `git commit -m 'Add amazing feature'`
6. **Push to the branch**: `git push origin feature/amazing-feature`
7. **Open a Pull Request**

### Contribution Guidelines

- Follow Go coding standards
- Add tests for new features
- Update documentation as needed
- Ensure all tests pass before submitting

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **[tview](https://github.com/rivo/tview)** - Terminal UI framework
- **[tcell](https://github.com/gdamore/tcell)** - Terminal cell manipulation
- **[Beep](https://github.com/faiface/beep)** - Audio playback library
- **[Oto](https://github.com/ebitengine/oto)** - Cross-platform audio library
- **Go community** - Excellent libraries and documentation

## 📊 Project Status

- ✅ **Core Features**: Audio playback, lyrics sync, visualizer
- ✅ **Cross-Platform**: Windows, macOS, Linux support
- ✅ **Local Development**: Complete build system and scripts
- ✅ **Documentation**: Comprehensive guides and examples
- 🔄 **Active Development**: Ready for GitHub publication and distribution

## 🌟 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=tuneminal/tuneminal&type=Date)](https://star-history.com/#tuneminal/tuneminal&Date)

---

**🎤 Happy Karaoke with Tuneminal! 🎵✨**

*Transform your terminal into a karaoke machine and sing your heart out!*
