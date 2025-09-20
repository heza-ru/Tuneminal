# 🎤 Tuneminal Jukebox

**Tuneminal** is an interactive command-line jukebox karaoke machine with live audio visualization built in Go. Experience a rich, animated interface that looks like a real jukebox - complete with song selection, synchronized lyrics, and stunning visual effects - all from your terminal!

![Tuneminal Demo](https://img.shields.io/badge/Go-1.21+-blue)
![License](https://img.shields.io/badge/License-MIT-green)
![Platform](https://img.shields.io/badge/Platform-Cross--Platform-lightgrey)

## ✨ Features

- 🎵 **Interactive Jukebox Interface**: Rich, animated terminal UI that looks like a real jukebox
- 🎨 **Multiple Themes**: Choose from Classic, Neon, Retro, Ocean, Forest, and Sunset themes
- 🎵 **Audio Playback**: Support for MP3 and WAV files using the Beep audio library
- 📝 **Synchronized Lyrics**: LRC format support with real-time scrolling and highlighting
- 🎨 **Live Visualizer**: Stunning animated bar-style visualization with particle effects
- 🖥️ **Tabbed Interface**: Organized tabs for Songs, Now Playing, and Settings
- 🔍 **Search Functionality**: Quick song search with real-time filtering
- ✨ **Smooth Animations**: Beautiful transitions and visual effects
- ⌨️ **Intuitive Controls**: Easy navigation with keyboard shortcuts
- 🎯 **Production Ready**: Well-structured, testable codebase

## 🚀 Installation

### Prerequisites

1. **Install Go** (version 1.21 or higher):
   - Windows: Download from [golang.org](https://golang.org/dl/)
   - macOS: `brew install go`
   - Linux: `sudo apt install golang-go` (Ubuntu/Debian)

2. **Verify Installation**:
   ```bash
   go version
   ```

### Build from Source

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
   go build -o tuneminal cmd/tuneminal/main.go
   ```

4. **Run Tuneminal**:
   ```bash
   ./tuneminal
   ```

## 🎮 Usage

### Getting Started

1. **Launch Tuneminal Jukebox**:
   ```bash
   ./tuneminal
   ```

2. **Navigate the Interface**: Use Tab to switch between tabs (Songs, Now Playing, Settings)
3. **Select a Song**: In the Songs tab, use arrow keys to navigate and Enter to play
4. **Enjoy Karaoke**: Watch the lyrics scroll and enjoy the visual effects!

### Controls

#### General Controls
- `Tab`: Switch between tabs (Songs, Now Playing, Settings)
- `Q`: Quit application
- `T`: Cycle through themes
- `A`: Toggle animations on/off

#### Song Selection
- `↑/↓` or `j/k`: Navigate songs
- `Enter`: Play selected song
- `/`: Open search (type to search, Esc to close)
- `R`: Refresh song list

#### Playback Controls
- `Space` or `Enter`: Play/Pause
- `S`: Stop playback
- `N`: Next song
- `P`: Previous song

### File Organization

Place your audio and lyrics files in the `uploads/demo/` directory:

```
uploads/
└── demo/
    ├── my_song.mp3
    ├── my_song.lrc
    ├── another_track.wav
    └── another_track.lrc
```

### LRC Lyrics Format

Create `.lrc` files with time-coded lyrics:

```lrc
[ar:Artist Name]
[ti:Song Title]
[al:Album Name]

[00:00.50]First line of lyrics
[00:03.25]Second line of lyrics
[00:06.00]Third line of lyrics
```

**Time Format**: `[mm:ss.xx]` where `xx` represents centiseconds (1/100th of a second)

## 🏗️ Project Structure

```
tuneminal/
├── cmd/
│   └── tuneminal/
│       └── main.go          # Application entry point
├── pkg/
│   ├── ui/                  # User interface components
│   │   ├── app.go          # Main application model
│   │   ├── menu.go         # File selection menu
│   │   ├── playback.go     # Karaoke playback view
│   │   └── visualizer.go   # Audio visualizer
│   ├── player/             # Audio playback engine
│   │   ├── player.go       # Audio player implementation
│   │   └── lyrics.go       # LRC lyrics parser
│   └── utils/              # Utility functions
│       └── files.go        # File system helpers
├── uploads/
│   └── demo/               # Demo files directory
├── go.mod                  # Go module definition
└── README.md              # This file
```

## 🛠️ Development

### Dependencies

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)**: Terminal UI framework
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)**: Styling and layout
- **[Beep](https://github.com/faiface/beep)**: Audio playback library

### Building

```bash
# Development build
go build -o tuneminal cmd/tuneminal/main.go

# Release build (optimized)
go build -ldflags="-s -w" -o tuneminal cmd/tuneminal/main.go
```

### Testing

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

## 🎵 Demo

Try the included demo song and lyrics:

1. The demo files are already included in `uploads/demo/`
2. Launch Tuneminal and select "demo_song.mp3"
3. Choose "demo_song.lrc" for lyrics
4. Enjoy the karaoke experience!

**Note**: The demo MP3 is a placeholder. Replace it with an actual audio file for full functionality.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Setup

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Add tests if applicable
5. Commit your changes: `git commit -m 'Add amazing feature'`
6. Push to the branch: `git push origin feature/amazing-feature`
7. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Charm Bracelet](https://charm.sh/) for the amazing terminal UI tools
- [Beep](https://github.com/faiface/beep) for audio playback capabilities
- The Go community for excellent libraries and documentation

## 🐛 Troubleshooting

### Common Issues

1. **"go: command not found"**
   - Install Go from [golang.org](https://golang.org/dl/)
   - Ensure Go is in your PATH

2. **Audio not playing**
   - Check that your audio files are in supported formats (MP3, WAV)
   - Verify file permissions
   - Ensure audio drivers are working

3. **Lyrics not syncing**
   - Check LRC file format (must use `[mm:ss.xx]` format)
   - Ensure lyrics file is in the same directory as audio file
   - Verify time codes are in chronological order

4. **Visualizer not showing**
   - Ensure terminal supports Unicode characters
   - Try resizing terminal window
   - Check that audio file is playing

### Getting Help

- Open an issue on GitHub
- Check existing issues for solutions
- Review the documentation above

---

**Happy Karaoke! 🎤✨**
