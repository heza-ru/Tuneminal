package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tuneminal/tuneminal/pkg/config"
	"github.com/tuneminal/tuneminal/pkg/export"
	"github.com/tuneminal/tuneminal/pkg/lyrics"
	"github.com/tuneminal/tuneminal/pkg/metadata"
	"github.com/tuneminal/tuneminal/pkg/player"
	"github.com/tuneminal/tuneminal/pkg/playlist"
)

// App represents the main Tuneminal application
type App struct {
	app           *tview.Application
	pages         *tview.Pages
	
	// Core components
	header        *tview.TextView
	songList      *tview.List
	nowPlaying    *tview.TextView
	visualizer    *tview.TextView
	statusBar     *tview.TextView
	searchInput   *tview.InputField
	lyrics        *tview.TextView
	progress      *tview.TextView
	score         *tview.TextView
	
	// Preloader
	preloader     *tview.TextView
	
	// Audio player
	player        *player.AudioPlayer

	// Configuration
	appConfig     *config.Config

	// Playlist management
	playlistManager *playlist.PlaylistManager
	currentPlaylist string

	// Lyrics editor
	lyricsEditor    *lyrics.LyricEditor

	// Export/Import
	exportManager   *export.ExportManager

	// State
	songs         []Song
	currentSong   int
	isPlaying     bool
	isPaused      bool
	position      time.Duration
	duration      time.Duration
	isLoading     bool // Prevent multiple simultaneous play attempts
	
	// Karaoke features
	lyricLines    []LyricLine
	karaokeScore  int
	streak        int
	accuracy      float64
	totalLyrics   int
	hitLyrics     int
	
	// Visualizer state
	visualizerBars []int
	beatPhase      int
	spectrumColors []string

	// Audio control state
	volume         float64
	shuffleMode    bool
	repeatMode     bool
	
	// Thread safety (simplified for stability)
	// stateMutex     sync.RWMutex
	
	// App state
	showPreloader bool
	preloaderDone bool
}

// Song represents a song in the library
type Song struct {
	Title      string
	Artist     string
	Path       string
	LyricsPath string
	Duration   time.Duration
}

// LyricLine represents a single line of lyrics with timing
type LyricLine struct {
	Time    time.Duration
	Text    string
	Index   int
	IsActive bool
	IsHit   bool
}

// NewApp creates a new Tuneminal application
func NewApp() *App {
	// Load configuration
	appConfig, err := config.LoadConfig(config.GetConfigPath())
	if err != nil {
		// Use default config if loading fails
		appConfig = config.DefaultConfig()
	}

	// Initialize audio player, playlist manager, lyrics editor, and export manager
	audioPlayer := player.NewAudioPlayer()
	playlistManager := playlist.NewPlaylistManager()
	lyricsEditor := lyrics.NewLyricEditor()
	exportManager := export.NewExportManager()

	app := &App{
		app:           tview.NewApplication(),
		player:        audioPlayer,
		appConfig:     appConfig,
		playlistManager: playlistManager,
		lyricsEditor:  lyricsEditor,
		exportManager: exportManager,
		songs:         []Song{},
		currentSong:   -1,
		showPreloader: true,
		preloaderDone: false,
		karaokeScore:  0,
		streak:        0,
		accuracy:      0.0,
		totalLyrics:   0,
		hitLyrics:     0,
		visualizerBars: make([]int, 12), // 12 frequency bands
		beatPhase:     0,
		spectrumColors: []string{"[red]", "[yellow]", "[green]", "[cyan]", "[blue]", "[magenta]"},
		volume:        appConfig.DefaultVolume,
		shuffleMode:   appConfig.ShuffleMode,
		repeatMode:    appConfig.RepeatMode,
	}
	
	app.setupUI()
	app.loadSongs()
	
	return app
}

// setupUI creates the user interface
func (a *App) setupUI() {
	// Create main pages container
	a.pages = tview.NewPages()
	
	// Create preloader page
	a.createPreloaderPage()
	
	// Create main application page
	a.createMainPage()
	
	// Set up key bindings
	a.setupKeyBindings()
	
	// Start with preloader
	a.app.SetRoot(a.pages, true)
	
	// Start preloader animation
	go a.preloaderAnimation()
}

// createPreloaderPage creates the preloader page
func (a *App) createPreloaderPage() {
	a.preloader = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(false)
	
	a.preloader.SetTextAlign(tview.AlignCenter)
	a.preloader.SetBorder(false)
	
	// Add preloader page
	a.pages.AddPage("preloader", a.preloader, true, true)
}

// createMainPage creates the main application page
func (a *App) createMainPage() {
	// Create all components
	a.createAllComponents()
	
	// Create main layout
	mainLayout := a.createMainLayout()
	
	// Add main page
	a.pages.AddPage("main", mainLayout, true, false)
}

// createAllComponents creates all UI components
func (a *App) createAllComponents() {
	// Header
	a.header = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(false)
	a.header.SetTextAlign(tview.AlignCenter)
	a.header.SetBorder(false)
	
	// Search input
	a.searchInput = tview.NewInputField().
		SetLabel("[cyan]Search: [white]").
		SetFieldWidth(25).
		SetChangedFunc(a.onSearchChanged)
	a.searchInput.SetBorder(true).
		SetTitle("[blue]Search Songs[white]").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(tcell.ColorBlue)
	
	// Add input capture for search field
	a.searchInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			// Clear search and return to song list
			a.searchInput.SetText("")
			a.filterAndUpdateSongList("")
			a.app.SetFocus(a.songList)
			return nil
		} else if event.Key() == tcell.KeyEnter {
			// Move to song list to select filtered results
			a.app.SetFocus(a.songList)
			return nil
		} else if event.Key() == tcell.KeyTab {
			// Tab exits search and returns focus to song list
			a.app.SetFocus(a.songList)
			return nil
		} else if event.Key() == tcell.KeyRune && event.Rune() == '/' {
			// '/' exits search and returns focus to song list
			a.app.SetFocus(a.songList)
			return nil
		}
		return event
	})
	
	// Song list
	a.songList = tview.NewList()
	a.songList.SetBorder(true).
		SetTitle("[yellow]Music Library[white]").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(tcell.ColorYellow)
	a.songList.SetSelectedBackgroundColor(tcell.ColorDarkBlue).
		SetSelectedTextColor(tcell.ColorWhite)
	
	// Now playing
	a.nowPlaying = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true)
	a.nowPlaying.SetBorder(true).
		SetTitle("[green]Now Playing[white]").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(tcell.ColorGreen)
	
	// Visualizer
	a.visualizer = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(false)
	a.visualizer.SetBorder(true).
		SetTitle("[magenta]Audio Visualizer[white]").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(tcell.ColorPurple)
	
	// Progress bar
	a.progress = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(false)
	a.progress.SetBorder(false)
	a.progress.SetTextAlign(tview.AlignCenter)
	
	// Controls section removed - instructions moved to help window (press 'h')
	
	// Lyrics with karaoke highlighting
	a.lyrics = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(false).
		SetTextAlign(tview.AlignCenter)
	a.lyrics.SetBorder(true).
		SetTitle("[red]Karaoke Lyrics[white]").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorRed)
	
	// Score display
	a.score = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(false)
	a.score.SetBorder(true).
		SetTitle("[yellow]Score[white]").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(tcell.ColorYellow)
	
	// Status bar
	a.statusBar = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(false)
	a.statusBar.SetBorder(false)
	a.statusBar.SetTextAlign(tview.AlignLeft)
}

// createMainLayout creates the main layout
func (a *App) createMainLayout() *tview.Flex {
	// Create main vertical layout
	mainLayout := tview.NewFlex().SetDirection(tview.FlexRow)
	
	// Header
	mainLayout.AddItem(a.header, 10, 1, false)
	
	// Search bar
	mainLayout.AddItem(a.searchInput, 3, 1, false)
	
	// Main content area (horizontal)
	contentArea := tview.NewFlex().SetDirection(tview.FlexColumn)
	
	// Left panel (songs + score)
	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	leftPanel.AddItem(a.songList, 0, 1, true)
	leftPanel.AddItem(a.score, 6, 1, false)
	contentArea.AddItem(leftPanel, 0, 1, true)
	
	// Right panel (now playing + visualizer)
	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	rightPanel.AddItem(a.nowPlaying, 0, 1, false)
	rightPanel.AddItem(a.visualizer, 0, 1, false)
	contentArea.AddItem(rightPanel, 0, 1, false)
	
	// Add content to main layout
	mainLayout.AddItem(contentArea, 0, 1, true)
	
	// Progress bar
	mainLayout.AddItem(a.progress, 1, 1, false)
	
	// Bottom panel (lyrics only - controls moved to help window)
	mainLayout.AddItem(a.lyrics, 0, 1, false)
	
	// Status bar
	mainLayout.AddItem(a.statusBar, 1, 1, false)
	
	return mainLayout
}

// setupKeyBindings sets up comprehensive key bindings
func (a *App) setupKeyBindings() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Check if help modal is open - if so, let it handle all input
		if a.pages.HasPage("help") {
			return event // Let the help modal handle input
		}
		
		// Check if search input has focus - if so, let it handle Tab and '/' normally
		currentFocus := a.app.GetFocus()
		if currentFocus == a.searchInput {
			// Only handle global quit commands when search has focus
			switch event.Key() {
			case tcell.KeyCtrlC:
				a.quit()
				return nil
			}
			// Let search input handle everything else (including Tab, '/', ESC, Enter)
			return event
		}
		
		switch event.Key() {
		case tcell.KeyCtrlC, tcell.KeyEscape:
			a.quit()
			return nil
		case tcell.KeyUp:
			a.navigateUp()
			return nil
		case tcell.KeyDown:
			a.navigateDown()
			return nil
		case tcell.KeyEnter:
			a.playSelectedSong()
			return nil
		case tcell.KeyTab:
			// Tab to switch between search and song list (only when search doesn't have focus)
			a.app.SetFocus(a.searchInput)
			return nil
		case tcell.KeyRight:
			a.seekForward()
			return nil
		case tcell.KeyLeft:
			a.seekBackward()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				a.quit()
				return nil
			case ' ':
				a.togglePlayPause()
				return nil
			case 's':
				a.stop()
				return nil
			case 'S':
				a.toggleShuffle()
				return nil
			case 'n':
				a.next()
				return nil
			case 'p':
				a.previous()
				return nil
			case '/':
				a.app.SetFocus(a.searchInput)
				return nil
			case 'r':
				a.loadSongs()
				return nil
			case 'R':
				a.toggleRepeat()
				return nil
			case 'l':
				a.app.SetFocus(a.lyrics)
				return nil
			case 'h':
				a.showHelp()
				return nil
			case '+':
				a.increaseVolume()
				return nil
			case '-':
				a.decreaseVolume()
				return nil
			case 'e':
				a.openLyricsEditor()
				return nil
			case 'f':
				a.showFileManager()
				return nil
			case 'x':
				a.showExportDialog()
				return nil
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				// Quick song selection - jump to song number
				songIndex := int(event.Rune() - '1')
				if songIndex >= 0 && songIndex < len(a.songs) {
					a.currentSong = songIndex
					a.updateSongList()
					a.updateNowPlaying()
					a.updateKaraokeLyrics()
					a.app.SetFocus(a.songList)
				}
				return nil
			case '0':
				// Jump to last song
				if len(a.songs) > 0 {
					a.currentSong = len(a.songs) - 1
					a.updateSongList()
					a.updateNowPlaying()
					a.updateKaraokeLyrics()
					a.app.SetFocus(a.songList)
				}
				return nil
			case 'v':
				// Quick volume toggle (mute/unmute)
				if a.volume > 0 {
					a.volume = 0
				} else {
					a.volume = 1.0
				}
				if a.player != nil {
					a.player.SetVolume(a.volume)
				}
				a.updateNowPlaying()
				a.saveConfig()
				return nil
			case 'm':
				// Mark current song as favorite (could extend for rating system)
				if a.currentSong >= 0 && a.currentSong < len(a.songs) {
					a.showMessage("⭐ Song marked as favorite!")
				}
				return nil
			case 'j':
				// Jump to specific time (show time input dialog)
				if a.isPlaying && a.currentSong >= 0 {
					a.showJumpToTimeDialog()
				}
				return nil
			case 'i':
				// Show song information
				if a.currentSong >= 0 && a.currentSong < len(a.songs) {
					a.showSongInfo()
				}
				return nil
			case 'k':
				// Toggle karaoke mode (hide/show lyrics during playback)
				a.toggleKaraokeDisplay()
				return nil
			case 'c':
				// Clear all scores and start fresh
				a.karaokeScore = 0
				a.streak = 0
				a.accuracy = 0.0
				a.hitLyrics = 0
				a.totalLyrics = 0
				a.updateScore()
				a.showMessage("🎯 Scores cleared!")
				return nil
			}
		}
		return event
	})
}

// preloaderAnimation runs the preloader animation
func (a *App) preloaderAnimation() {
	steps := []string{
		`[white]╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║ [yellow]████████╗██╗   ██╗███╗   ██╗███████╗███╗   ███╗██╗███╗   ██╗ █████╗ ██╗     [white] ║
║ [yellow]╚══██╔══╝██║   ██║████╗  ██║██╔════╝████╗ ████║██║████╗  ██║██╔══██╗██║     [white] ║
║ [yellow]   ██║   ██║   ██║██╔██╗ ██║█████╗  ██╔████╔██║██║██╔██╗ ██║███████║██║     [white] ║
║ [yellow]   ██║   ██║   ██║██║╚██╗██║██╔══╝  ██║╚██╔╝██║██║██║╚██╗██║██╔══██║██║     [white] ║
║ [yellow]   ██║   ╚██████╔╝██║ ╚████║███████╗██║ ╚═╝ ██║██║██║ ╚████║██║  ██║███████╗[white] ║
║ [yellow]   ╚═╝    ╚═════╝ ╚═╝  ╚═══╝╚══════╝╚═╝     ╚═╝╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝╚══════╝[white] ║
║                                                                              ║
║                            [cyan]KARAOKE MACHINE[white]                            ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝

[cyan]Initializing audio system...[white]`,
		`[white]╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║ [yellow]████████╗██╗   ██╗███╗   ██╗███████╗███╗   ███╗██╗███╗   ██╗ █████╗ ██╗     [white] ║
║ [yellow]╚══██╔══╝██║   ██║████╗  ██║██╔════╝████╗ ████║██║████╗  ██║██╔══██╗██║     [white] ║
║ [yellow]   ██║   ██║   ██║██╔██╗ ██║█████╗  ██╔████╔██║██║██╔██╗ ██║███████║██║     [white] ║
║ [yellow]   ██║   ██║   ██║██║╚██╗██║██╔══╝  ██║╚██╔╝██║██║██║╚██╗██║██╔══██║██║     [white] ║
║ [yellow]   ██║   ╚██████╔╝██║ ╚████║███████╗██║ ╚═╝ ██║██║██║ ╚████║██║  ██║███████╗[white] ║
║ [yellow]   ╚═╝    ╚═════╝ ╚═╝  ╚═══╝╚══════╝╚═╝     ╚═╝╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝╚══════╝[white] ║
║                                                                              ║
║                            [cyan]KARAOKE MACHINE[white]                            ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝

[cyan]Loading music library...[white]
[green]✓[white] Audio system ready`,
		`[white]╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║ [yellow]████████╗██╗   ██╗███╗   ██╗███████╗███╗   ███╗██╗███╗   ██╗ █████╗ ██╗     [white] ║
║ [yellow]╚══██╔══╝██║   ██║████╗  ██║██╔════╝████╗ ████║██║████╗  ██║██╔══██╗██║     [white] ║
║ [yellow]   ██║   ██║   ██║██╔██╗ ██║█████╗  ██╔████╔██║██║██╔██╗ ██║███████║██║     [white] ║
║ [yellow]   ██║   ██║   ██║██║╚██╗██║██╔══╝  ██║╚██╔╝██║██║██║╚██╗██║██╔══██║██║     [white] ║
║ [yellow]   ██║   ╚██████╔╝██║ ╚████║███████╗██║ ╚═╝ ██║██║██║ ╚████║██║  ██║███████╗[white] ║
║ [yellow]   ╚═╝    ╚═════╝ ╚═╝  ╚═══╝╚══════╝╚═╝     ╚═╝╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝╚══════╝[white] ║
║                                                                              ║
║                            [cyan]KARAOKE MACHINE[white]                            ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝

[cyan]Setting up karaoke features...[white]
[green]✓[white] Audio system ready
[green]✓[white] Music library loaded`,
		`[white]╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║ [yellow]████████╗██╗   ██╗███╗   ██╗███████╗███╗   ███╗██╗███╗   ██╗ █████╗ ██╗     [white] ║
║ [yellow]╚══██╔══╝██║   ██║████╗  ██║██╔════╝████╗ ████║██║████╗  ██║██╔══██╗██║     [white] ║
║ [yellow]   ██║   ██║   ██║██╔██╗ ██║█████╗  ██╔████╔██║██║██╔██╗ ██║███████║██║     [white] ║
║ [yellow]   ██║   ██║   ██║██║╚██╗██║██╔══╝  ██║╚██╔╝██║██║██║╚██╗██║██╔══██║██║     [white] ║
║ [yellow]   ██║   ╚██████╔╝██║ ╚████║███████╗██║ ╚═╝ ██║██║██║ ╚████║██║  ██║███████╗[white] ║
║ [yellow]   ╚═╝    ╚═════╝ ╚═╝  ╚═══╝╚══════╝╚═╝     ╚═╝╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝╚══════╝[white] ║
║                                                                              ║
║                            [cyan]KARAOKE MACHINE[white]                            ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝

[cyan]Preparing real-time lyrics...[white]
[green]✓[white] Audio system ready
[green]✓[white] Music library loaded
[green]✓[white] Karaoke features ready`,
		`[white]╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║ [yellow]████████╗██╗   ██╗███╗   ██╗███████╗███╗   ███╗██╗███╗   ██╗ █████╗ ██╗     [white] ║
║ [yellow]╚══██╔══╝██║   ██║████╗  ██║██╔════╝████╗ ████║██║████╗  ██║██╔══██╗██║     [white] ║
║ [yellow]   ██║   ██║   ██║██╔██╗ ██║█████╗  ██╔████╔██║██║██╔██╗ ██║███████║██║     [white] ║
║ [yellow]   ██║   ██║   ██║██║╚██╗██║██╔══╝  ██║╚██╔╝██║██║██║╚██╗██║██╔══██║██║     [white] ║
║ [yellow]   ██║   ╚██████╔╝██║ ╚████║███████╗██║ ╚═╝ ██║██║██║ ╚████║██║  ██║███████╗[white] ║
║ [yellow]   ╚═╝    ╚═════╝ ╚═╝  ╚═══╝╚══════╝╚═╝     ╚═╝╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝╚══════╝[white] ║
║                                                                              ║
║                            [cyan]KARAOKE MACHINE[white]                            ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝

[cyan]Starting karaoke session...[white]
[green]✓[white] Audio system ready
[green]✓[white] Music library loaded
[green]✓[white] Karaoke features ready
[green]✓[white] Real-time lyrics ready

[yellow]Ready to sing![white]`,
	}
	
	for i, step := range steps {
		a.app.QueueUpdateDraw(func() {
			a.preloader.SetText(step)
		})
		time.Sleep(500 * time.Millisecond)
		
		// After the last step, switch to main page
		if i == len(steps)-1 {
			time.Sleep(1 * time.Second)
			a.app.QueueUpdateDraw(func() {
				a.pages.SwitchToPage("main")
				a.showPreloader = false
				a.preloaderDone = true
				a.updateAllDisplays()
				// Force focus to song list
				a.app.SetFocus(a.songList)
			})
		}
	}
}

// findLyricsFile finds the corresponding lyrics file for an audio file
func (a *App) findLyricsFile(audioPath string) string {
	// Replace audio extension with .lrc
	ext := filepath.Ext(audioPath)
	lyricsPath := strings.TrimSuffix(audioPath, ext) + ".lrc"
	
	// Check if lyrics file exists
	if _, err := os.Stat(lyricsPath); err == nil {
		return lyricsPath
	}
	
	// Try alternative naming patterns
	baseName := strings.TrimSuffix(filepath.Base(audioPath), ext)
	dir := filepath.Dir(audioPath)
	
	// Try different patterns
	patterns := []string{
		filepath.Join(dir, baseName + ".lrc"),
		filepath.Join(dir, strings.ReplaceAll(baseName, "_", " ") + ".lrc"),
		filepath.Join(dir, strings.ReplaceAll(baseName, "_", "-") + ".lrc"),
	}
	
	for _, pattern := range patterns {
		if _, err := os.Stat(pattern); err == nil {
			return pattern
		}
	}
	
	return ""
}

// loadSongs loads songs with real metadata from files
func (a *App) loadSongs() {
	// Scan directory for real audio files with metadata
	songMetadata, err := metadata.ScanDirectory("uploads/demo")
	if err != nil {
		return
	}
	
	// Convert metadata to app songs
	a.songs = []Song{}
	
	for _, meta := range songMetadata {
		appSong := Song{
			Title:      meta.Title,
			Artist:     meta.Artist,
			Path:       meta.Path,
			LyricsPath: a.findLyricsFile(meta.Path),
			Duration:   meta.Duration,
		}
		a.songs = append(a.songs, appSong)
	}
	
	// Set default selection to first song if available
	if len(a.songs) > 0 {
	a.currentSong = 0
	}
	
	// Update displays
	a.updateAllDisplays()
}

// loadDemoLyrics loads demo lyrics with timing
func (a *App) loadDemoLyrics() {
	a.lyricLines = []LyricLine{
		{Time: 0 * time.Second, Text: "Welcome to Tuneminal Karaoke!", Index: 0, IsActive: false, IsHit: false},
		{Time: 2 * time.Second, Text: "", Index: 1, IsActive: false, IsHit: false},
		{Time: 3 * time.Second, Text: "This is a demo song", Index: 2, IsActive: false, IsHit: false},
		{Time: 5 * time.Second, Text: "For the Tuneminal app", Index: 3, IsActive: false, IsHit: false},
		{Time: 7 * time.Second, Text: "Karaoke in your terminal", Index: 4, IsActive: false, IsHit: false},
		{Time: 9 * time.Second, Text: "It's really quite a snap", Index: 5, IsActive: false, IsHit: false},
		{Time: 11 * time.Second, Text: "", Index: 6, IsActive: false, IsHit: false},
		{Time: 12 * time.Second, Text: "Chorus", Index: 7, IsActive: false, IsHit: false},
		{Time: 14 * time.Second, Text: "Sing along with me", Index: 8, IsActive: false, IsHit: false},
		{Time: 16 * time.Second, Text: "In your terminal today", Index: 9, IsActive: false, IsHit: false},
		{Time: 18 * time.Second, Text: "Tuneminal makes it easy", Index: 10, IsActive: false, IsHit: false},
		{Time: 20 * time.Second, Text: "To karaoke the Go way", Index: 11, IsActive: false, IsHit: false},
		{Time: 22 * time.Second, Text: "", Index: 12, IsActive: false, IsHit: false},
		{Time: 23 * time.Second, Text: "Verse 2", Index: 13, IsActive: false, IsHit: false},
		{Time: 25 * time.Second, Text: "No need for fancy GUIs", Index: 14, IsActive: false, IsHit: false},
		{Time: 27 * time.Second, Text: "Just terminal and text", Index: 15, IsActive: false, IsHit: false},
		{Time: 29 * time.Second, Text: "Tuneminal brings the music", Index: 16, IsActive: false, IsHit: false},
		{Time: 31 * time.Second, Text: "To your command line next", Index: 17, IsActive: false, IsHit: false},
		{Time: 33 * time.Second, Text: "", Index: 18, IsActive: false, IsHit: false},
		{Time: 34 * time.Second, Text: "Thank you for using Tuneminal!", Index: 19, IsActive: false, IsHit: false},
	}
}

// loadLyricsFromFile loads lyrics from an LRC file
func (a *App) loadLyricsFromFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		// If file doesn't exist, use demo lyrics
		a.loadDemoLyrics()
		return
	}
	defer file.Close()

	a.lyricLines = []LyricLine{}
	scanner := bufio.NewScanner(file)
	index := 0
	
	// Regex to match LRC time format [mm:ss.xx]
	timeRegex := regexp.MustCompile(`\[(\d{2}):(\d{2})\.(\d{2})\]`)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and metadata lines
		if line == "" || strings.HasPrefix(line, "[") && !timeRegex.MatchString(line) {
			continue
		}
		
		// Extract time and text
		matches := timeRegex.FindStringSubmatch(line)
		if len(matches) == 4 {
			minutes, _ := strconv.Atoi(matches[1])
			seconds, _ := strconv.Atoi(matches[2])
			centiseconds, _ := strconv.Atoi(matches[3])
			
			time := time.Duration(minutes)*time.Minute + 
					time.Duration(seconds)*time.Second + 
					time.Duration(centiseconds)*10*time.Millisecond
			
			// Extract text after the time tag
			text := timeRegex.ReplaceAllString(line, "")
			text = strings.TrimSpace(text)
			
			a.lyricLines = append(a.lyricLines, LyricLine{
				Time:     time,
				Text:     text,
				Index:    index,
				IsActive: false,
				IsHit:    false,
			})
			index++
		}
	}
	
	// If no lyrics were loaded, use demo lyrics
	if len(a.lyricLines) == 0 {
		a.loadDemoLyrics()
	}
}

// updateAllDisplays updates all display components
func (a *App) updateAllDisplays() {
	a.updateHeader()
	a.updateSongList()
	a.updateNowPlaying()
	a.updateProgress()
	a.updateKaraokeLyrics()
	a.updateScore()
	a.updateStatus()
}

// updateHeader updates the header display
func (a *App) updateHeader() {
	title := `[white]╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║ [yellow]████████╗██╗   ██╗███╗   ██╗███████╗███╗   ███╗██╗███╗   ██╗ █████╗ ██╗     [white] ║
║ [yellow]╚══██╔══╝██║   ██║████╗  ██║██╔════╝████╗ ████║██║████╗  ██║██╔══██╗██║     [white] ║
║ [yellow]   ██║   ██║   ██║██╔██╗ ██║█████╗  ██╔████╔██║██║██╔██╗ ██║███████║██║     [white] ║
║ [yellow]   ██║   ██║   ██║██║╚██╗██║██╔══╝  ██║╚██╔╝██║██║██║╚██╗██║██╔══██║██║     [white] ║
║ [yellow]   ██║   ╚██████╔╝██║ ╚████║███████╗██║ ╚═╝ ██║██║██║ ╚████║██║  ██║███████╗[white] ║
║ [yellow]   ╚═╝    ╚═════╝ ╚═╝  ╚═══╝╚══════╝╚═╝     ╚═╝╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝╚══════╝[white] ║
║                                                                              ║
║                            [cyan]KARAOKE MACHINE[white]                            ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝`
	
	a.header.SetText(title)
}

// updateSongList updates the song list display
func (a *App) updateSongList() {
	a.songList.Clear()
	
	for i, song := range a.songs {
		title := fmt.Sprintf("%s - %s [%s]", song.Title, song.Artist, formatDuration(song.Duration))
		
		// Add status prefix
		if i == a.currentSong {
			if a.isPlaying {
				if a.isPaused {
					title = "⏸ " + title
				} else {
					title = "▶ " + title
				}
			} else {
				title = "● " + title
			}
		} else {
			title = "  " + title
		}
		
		a.songList.AddItem(title, "", 0, nil)
	}
	
	// Set the current selection
	if a.currentSong >= 0 && a.currentSong < len(a.songs) {
		a.songList.SetCurrentItem(a.currentSong)
	}
}

// updateNowPlaying updates the now playing display
func (a *App) updateNowPlaying() {
	if a.currentSong < 0 || a.currentSong >= len(a.songs) {
		a.nowPlaying.SetText("[white]No song selected[white]")
		return
	}
	
	song := a.songs[a.currentSong]
	
	// Get current volume percentage
	volumePercent := int(a.volume * 100)

	// Create now playing text
	playlistInfo := ""
	if a.currentPlaylist != "" {
		playlistInfo = fmt.Sprintf("\n[white]Playlist: [cyan]%s[white]", a.currentPlaylist)
	}

	text := fmt.Sprintf(`[white]Title: [yellow]%s[white]
Artist: [yellow]%s[white]
Duration: [yellow]%s[white]
Position: [yellow]%s[white]

[white]Status: [green]%s[white]%s

[white]Volume: [cyan]%d%%[white]
[white]Repeat: [cyan]%s[white]
[white]Shuffle: [cyan]%s[white]`,
		song.Title,
		song.Artist,
		formatDuration(song.Duration),
		formatDuration(a.position),
		a.getStatusText(),
		playlistInfo,
		volumePercent,
		a.getRepeatModeText(),
		a.getShuffleModeText())
	
	a.nowPlaying.SetText(text)
}

// updateKaraokeLyrics creates a beautiful 5-line auto-scrolling karaoke display
func (a *App) updateKaraokeLyrics() {
	if len(a.lyricLines) == 0 {
		a.lyrics.SetText(a.createEmptyLyricsDisplay())
		return
	}
	
	// Find current active lyric line
	currentTime := a.position
	activeIndex := a.findCurrentLyricIndex(currentTime)
	
	// Create 5-line display with current line in center (index 2)
	display := a.createFiveLineLyricsDisplay(activeIndex)
	
	a.lyrics.SetText(display)
}

// findCurrentLyricIndex finds the index of the currently active lyric
func (a *App) findCurrentLyricIndex(currentTime time.Duration) int {
	activeIndex := -1
	
	for i, lyric := range a.lyricLines {
		if currentTime >= lyric.Time {
			activeIndex = i
		} else {
			break
		}
	}
	
	return activeIndex
}

// createFiveLineLyricsDisplay creates a beautiful 5-line karaoke display that ALWAYS shows 5 lines
func (a *App) createFiveLineLyricsDisplay(activeIndex int) string {
	var display strings.Builder
	
	// Add top padding for better vertical centering
	display.WriteString("\n")
	
	// ALWAYS create exactly 5 lines: 2 previous, 1 current, 2 upcoming
	lines := []string{}
	
	for line := 0; line < 5; line++ {
		lyricIndex := activeIndex + (line - 2) // Center current line at position 2
		var formattedLine string
		
		if line == 0 || line == 4 {
			// Top and bottom padding lines (very subtle)
			formattedLine = a.formatLyricLine(lyricIndex, "padding")
		} else if line == 1 {
			// Previous line (completed)
			formattedLine = a.formatLyricLine(lyricIndex, "previous")
		} else if line == 2 {
			// Current line (center) - MUCH LARGER and highlighted
			formattedLine = a.formatLyricLine(lyricIndex, "current")
		} else if line == 3 {
			// Next line (upcoming)
			formattedLine = a.formatLyricLine(lyricIndex, "next")
		}
		
		// Ensure each line has content (never empty)
		if formattedLine == "" {
			if line == 2 {
				formattedLine = a.formatEmptyLine("current")
			} else if line == 1 {
				formattedLine = a.formatEmptyLine("previous")
			} else if line == 3 {
				formattedLine = a.formatEmptyLine("next")
		} else {
				formattedLine = a.formatEmptyLine("padding")
			}
		}
		
		lines = append(lines, formattedLine)
	}
	
	// Build the display with guaranteed 5 lines and better vertical centering
	display.WriteString("\n\n")  // Top padding for vertical centering
	
	for i, line := range lines {
		display.WriteString(line)
		
		// Add consistent spacing between lines for better centering
		if i == 0 {
			display.WriteString("\n\n")   // Space after first padding line
		} else if i == 1 {
			display.WriteString("\n\n")   // Space before current line
		} else if i == 2 {
			display.WriteString("\n\n")   // Space after current line
		} else if i == 3 {
			display.WriteString("\n\n")   // Space after next line
		} else if i == 4 {
			display.WriteString("\n")     // Minimal space after last padding line
		}
	}
	
	// Add bottom padding for vertical centering
	display.WriteString("\n\n")
	
	return display.String()
}

// formatLyricLine formats a single lyric line based on its position and type
func (a *App) formatLyricLine(index int, lineType string) string {
	// Handle edge cases
	if index < 0 || index >= len(a.lyricLines) {
		return a.formatEmptyLine(lineType)
	}
	
	lyric := a.lyricLines[index]
	text := lyric.Text
	
	// Skip empty lyrics
	if text == "" {
		return a.formatEmptyLine(lineType)
	}
	
	switch lineType {
	case "current":
		// Current line: MUCH LARGER with animated beat indicators
		beatIndicator := "♪"
		if a.position.Milliseconds()%1000 < 500 {
			beatIndicator = "♫"
		}
		// Create a large, prominent display with uppercase text
		upperText := strings.ToUpper(text)
		return fmt.Sprintf("[yellow::b]%s  %s  %s[white::-]", beatIndicator, upperText, beatIndicator)
		
	case "previous":
		// Previous line: Smaller, completed style
		return fmt.Sprintf("[blue::d]%s[white::-]", text)
		
	case "next":
		// Next line: Normal size, upcoming
		return fmt.Sprintf("[white]%s", text)
		
	case "padding":
		// Padding lines: Very subtle but visible on all backgrounds
		if text != "" {
			return fmt.Sprintf("[gray]%s[white::-]", text)
		}
		return "[gray]∙∙∙[white::-]"
		
	default:
		return fmt.Sprintf("[white]%s", text)
	}
}

// formatEmptyLine formats an empty line based on its type
func (a *App) formatEmptyLine(lineType string) string {
	switch lineType {
	case "current":
		// Large, animated ready state
		beatIndicator := "♪"
		if time.Now().Unix()%2 == 0 {
			beatIndicator = "♫"
		}
		return fmt.Sprintf("[yellow::b]%s  READY TO SING  %s[white::-]", beatIndicator, beatIndicator)
	case "previous":
		return "[gray]∙∙∙[white::-]"
	case "next":
		return "[gray]∙∙∙[white::-]"
	case "padding":
		return "[gray] [white::-]"
	default:
		return "[gray]∙∙∙[white::-]"
	}
}

// createEmptyLyricsDisplay creates a display for when no lyrics are available
func (a *App) createEmptyLyricsDisplay() string {
	// Add some animation to the empty state
	beatIndicator := "♪"
	if time.Now().Unix()%2 == 0 {
		beatIndicator = "♫"
	}
	
	return fmt.Sprintf(`


[gray]∙∙∙[white::-]


[gray]∙∙∙[white::-]


[yellow::b]%s  NO LYRICS AVAILABLE  %s[white::-]


[gray]∙∙∙[white::-]


[gray]∙∙∙[white::-]


`, beatIndicator, beatIndicator)
}

// updateScore updates the dynamic scoring display
func (a *App) updateScore() {
	// Update scoring during playback
	if a.isPlaying {
		a.updateKaraokeScoring()
	}
	
	// Create dynamic score display
	scoreDisplay := a.createScoreDisplay()
	a.score.SetText(scoreDisplay)
}

// updateKaraokeScoring processes real-time karaoke scoring
func (a *App) updateKaraokeScoring() {
	if len(a.lyricLines) == 0 {
		return
	}
	
	// Find current active lyric
	currentTime := a.position
	activeIndex := a.findCurrentLyricIndex(currentTime)
	
	if activeIndex >= 0 && activeIndex < len(a.lyricLines) {
		lyric := &a.lyricLines[activeIndex]
		
		// Auto-hit system: simulate user singing along
		if !lyric.IsHit && !lyric.IsActive {
			// Mark as active when reached
			lyric.IsActive = true
			
			// Simulate singing performance (creative scoring)
			hitChance := a.calculateHitChance(activeIndex)
			if rand.Float64() < hitChance {
				a.hitLyric(activeIndex)
			}
		}
	}
	
	// Update accuracy
	a.accuracy = a.calculateAccuracy()
}

// calculateHitChance determines likelihood of hitting a lyric line
func (a *App) calculateHitChance(lyricIndex int) float64 {
	baseChance := 0.7 // 70% base hit rate
	
	// Bonus for streak
	streakBonus := float64(a.streak) * 0.05 // +5% per streak
	if streakBonus > 0.2 {
		streakBonus = 0.2 // Cap at +20%
	}
	
	// Song progress bonus (easier as song progresses)
	progressBonus := (float64(lyricIndex) / float64(len(a.lyricLines))) * 0.1
	
	// Beat synchronization bonus
	beatBonus := 0.0
	if a.beatPhase == 0 { // On beat
		beatBonus = 0.1
	}
	
	totalChance := baseChance + streakBonus + progressBonus + beatBonus
	if totalChance > 0.95 {
		totalChance = 0.95 // Cap at 95%
	}
	
	return totalChance
}

// hitLyric processes a successful lyric hit
func (a *App) hitLyric(lyricIndex int) {
	if lyricIndex < 0 || lyricIndex >= len(a.lyricLines) {
		return
	}
	
	lyric := &a.lyricLines[lyricIndex]
		if lyric.IsHit {
		return // Already hit
	}
	
	// Mark as hit
	lyric.IsHit = true
	a.hitLyrics++
	
	// Calculate points for this hit
	basePoints := 100
	streakMultiplier := 1.0 + (float64(a.streak) / 10.0) // +10% per streak level
	beatBonus := 0
	if a.beatPhase == 0 {
		beatBonus = 50 // Bonus for hitting on beat
	}
	
	points := int(float64(basePoints) * streakMultiplier) + beatBonus
	a.karaokeScore += points
	
	// Update streak
	a.streak++
	
	// Achievement bonuses
	if a.streak == 5 {
		a.karaokeScore += 500 // First streak bonus
	} else if a.streak == 10 {
		a.karaokeScore += 1000 // Perfect streak bonus
	} else if a.streak%15 == 0 {
		a.karaokeScore += 2000 // Legendary streak bonus
	}
}

// calculateAccuracy calculates current singing accuracy
func (a *App) calculateAccuracy() float64 {
	if a.totalLyrics == 0 {
		a.totalLyrics = len(a.lyricLines)
	}
	
	if a.totalLyrics == 0 {
		return 0.0
	}
	
	return float64(a.hitLyrics) / float64(a.totalLyrics) * 100.0
}

// createScoreDisplay builds the dynamic score display
func (a *App) createScoreDisplay() string {
	var display strings.Builder
	
	// Score with dynamic color based on performance
	scoreColor := a.getScoreColor()
	display.WriteString(fmt.Sprintf("%sScore: %d[white]\n", scoreColor, a.karaokeScore))
	
	// Streak with special effects
	streakDisplay := a.getStreakDisplay()
	display.WriteString(fmt.Sprintf("%s\n", streakDisplay))
	
	// Accuracy with performance indicator
	accuracyColor := a.getAccuracyColor()
	display.WriteString(fmt.Sprintf("%sAccuracy: %.1f%%[white]\n\n", accuracyColor, a.accuracy))
	
	// Dynamic status and achievements
	status := a.getPerformanceStatus()
	display.WriteString(fmt.Sprintf("%s\n\n", status))
	
	// Performance tips based on current state
	tips := a.getPerformanceTips()
	display.WriteString(tips)
	
	return display.String()
}

// getScoreColor returns color based on score level
func (a *App) getScoreColor() string {
	if a.karaokeScore >= 10000 {
		return "[red::b]" // Legendary
	} else if a.karaokeScore >= 5000 {
		return "[magenta::b]" // Amazing
	} else if a.karaokeScore >= 2000 {
		return "[yellow::b]" // Great
	} else if a.karaokeScore >= 1000 {
		return "[green]" // Good
			} else {
		return "[white]" // Starting out
	}
}

// getStreakDisplay creates dynamic streak display
func (a *App) getStreakDisplay() string {
	if a.streak == 0 {
		return "[dim]Streak: 0[white]"
	} else if a.streak < 5 {
		return fmt.Sprintf("[green]Streak: %d[white]", a.streak)
	} else if a.streak < 10 {
		return fmt.Sprintf("[yellow::b]Streak: %d[white]", a.streak)
	} else if a.streak < 20 {
		return fmt.Sprintf("[red::b]STREAK: %d[white]", a.streak)
	} else {
		return fmt.Sprintf("[magenta::b]LEGENDARY: %d[white]", a.streak)
	}
}

// getAccuracyColor returns color based on accuracy
func (a *App) getAccuracyColor() string {
	if a.accuracy >= 90 {
		return "[green::b]"
	} else if a.accuracy >= 75 {
		return "[yellow]"
	} else if a.accuracy >= 50 {
		return "[white]"
	} else {
		return "[dim]"
	}
}

// getPerformanceStatus returns dynamic status message
func (a *App) getPerformanceStatus() string {
	if !a.isPlaying {
		return "[cyan]🎤 Ready to Sing! 🎤[white]"
	}
	
	if a.streak >= 15 {
		return "[magenta::b]🌟 LEGENDARY PERFORMANCE! 🌟[white]"
	} else if a.streak >= 10 {
		return "[red::b]🔥 ON FIRE! UNSTOPPABLE! 🔥[white]"
	} else if a.streak >= 5 {
		return "[yellow::b]⚡ GREAT RHYTHM! KEEP GOING! ⚡[white]"
	} else if a.accuracy >= 80 {
		return "[green]🎵 Excellent Singing! 🎵[white]"
	} else if a.accuracy >= 60 {
		return "[white]🎶 Good Performance! 🎶[white]"
	} else {
		return "[cyan]🎤 Finding Your Rhythm... 🎤[white]"
	}
}

// getPerformanceTips returns contextual tips
func (a *App) getPerformanceTips() string {
	if !a.isPlaying {
		return `[white]Tips:
[cyan]•[white] Follow the highlighted lyrics
[cyan]•[white] Keep the beat for bonus points
[cyan]•[white] Build streaks for multipliers`
	}
	
	if a.streak == 0 {
		return `[yellow]Boost Your Score:
[cyan]•[white] Sing along with current line
[cyan]•[white] Stay in rhythm with the beat
[cyan]•[white] Hit consecutive lines for streaks`
	} else if a.streak < 5 {
		return `[green]Building Streak:
[cyan]•[white] Keep singing to maintain streak
[cyan]•[white] Watch for beat indicators
[cyan]•[white] 5-streak bonus coming up!`
	} else {
		return `[red::b]Streak Master:
[cyan]•[white] You're in the zone!
[cyan]•[white] Streak multiplier active
[cyan]•[white] Keep the energy high!`
	}
}

// updateVisualizer creates dynamic audio visualizations
func (a *App) updateVisualizer() {
	if !a.isPlaying {
		// Static display when not playing
		a.visualizer.SetText(`[white]♪ Audio Spectrum ♪

[dim]   ░░░░  ░░░░  ░░░░  ░░░░  ░░░░  ░░░░
   ░░░░  ░░░░  ░░░░  ░░░░  ░░░░  ░░░░
   ░░░░  ░░░░  ░░░░  ░░░░  ░░░░  ░░░░
   ░░░░  ░░░░  ░░░░  ░░░░  ░░░░  ░░░░
   ░░░░  ░░░░  ░░░░  ░░░░  ░░░░  ░░░░

  Bass   Low   Mid  High  Treble Ultra[white]`)
		return
	}
	
	// Dynamic visualization during playback
	a.generateVisualizerData()
	display := a.createVisualizerDisplay()
	a.visualizer.SetText(display)
}

// generateVisualizerData creates dynamic audio visualization data
func (a *App) generateVisualizerData() {
	// Simulate audio analysis with position-based patterns
	timeMs := a.position.Milliseconds()
	
	// Update beat phase for rhythm sync
	a.beatPhase = int(timeMs/250) % 4 // 4-beat pattern
	
	// Generate frequency band heights (0-8)
	for i := 0; i < len(a.visualizerBars); i++ {
		// Create different patterns for different frequency bands
		baseHeight := 2
		
		// Bass frequencies (0-2) - lower, pulsing with beat
		if i < 3 {
			beatBoost := 0
			if a.beatPhase == 0 || a.beatPhase == 2 {
				beatBoost = 3
			}
			a.visualizerBars[i] = baseHeight + beatBoost + rand.Intn(2)
		}
		// Mid frequencies (3-6) - more active
		if i >= 3 && i < 7 {
			a.visualizerBars[i] = baseHeight + 2 + rand.Intn(3)
		}
		// High frequencies (7-11) - most active, sparkly
		if i >= 7 {
			sparkle := rand.Intn(4)
			if rand.Float32() < 0.3 { // 30% chance of spike
				sparkle += 3
			}
			a.visualizerBars[i] = baseHeight + sparkle
		}
		
		// Ensure bars don't exceed maximum height
		if a.visualizerBars[i] > 8 {
			a.visualizerBars[i] = 8
		}
	}
}

// createVisualizerDisplay builds the visual representation
func (a *App) createVisualizerDisplay() string {
	var display strings.Builder
	
	// Title with beat indicator
	beatIndicator := "♪"
	if a.beatPhase%2 == 0 {
		beatIndicator = "♫"
	}
	display.WriteString(fmt.Sprintf("[white]%s Live Audio Spectrum %s\n\n", beatIndicator, beatIndicator))
	
	// Draw spectrum bars (8 rows, 12 columns)
	for row := 7; row >= 0; row-- { // Top to bottom
		display.WriteString("  ")
		for col := 0; col < 12; col++ {
			barHeight := a.visualizerBars[col]
			
			if row < barHeight {
				// Choose color based on frequency band and intensity
				color := a.getVisualizerColor(col, row, barHeight)
				display.WriteString(fmt.Sprintf("%s████[white]  ", color))
			} else {
				// Empty space or dim background
				if row == 0 {
					display.WriteString("[dim]░░░░[white]  ")
				} else {
					display.WriteString("      ")
				}
			}
		}
		display.WriteString("\n")
	}
	
	// Frequency labels
	display.WriteString("\n Bass  Low  Mid  Mid  High High Treb Treb  Ultr Ultr  Air  Air\n")
	
	// Add dynamic status
	intensity := a.calculateVisualizerIntensity()
	status := a.getIntensityStatus(intensity)
	display.WriteString(fmt.Sprintf("\n[white]%s[white]", status))
	
	return display.String()
}

// getVisualizerColor returns appropriate color for frequency band and intensity
func (a *App) getVisualizerColor(band, row, height int) string {
	// Color mapping based on frequency band
	if band < 3 { // Bass - red to yellow
		if row >= height-2 {
			return "[yellow::b]" // Peak
		}
		return "[red]"
	} else if band < 7 { // Mid - green to cyan
		if row >= height-2 {
			return "[cyan::b]" // Peak
		}
		return "[green]"
	} else { // High - blue to magenta
		if row >= height-2 {
			return "[magenta::b]" // Peak
		}
		return "[blue]"
	}
}

// calculateVisualizerIntensity gets overall audio intensity
func (a *App) calculateVisualizerIntensity() float64 {
	total := 0
	for _, bar := range a.visualizerBars {
		total += bar
	}
	return float64(total) / float64(len(a.visualizerBars)*8) // Normalize to 0-1
}

// getIntensityStatus returns status message based on intensity
func (a *App) getIntensityStatus(intensity float64) string {
	if intensity > 0.8 {
		return "[red::b]🔥 INTENSE ENERGY! 🔥[white]"
	} else if intensity > 0.6 {
		return "[yellow::b]⚡ HIGH ENERGY ⚡[white]"
	} else if intensity > 0.4 {
		return "[green]🎵 MODERATE VIBES 🎵[white]"
	} else if intensity > 0.2 {
		return "[cyan]~ Gentle Flow ~[white]"
	} else {
		return "[dim]∙ Quiet Moment ∙[white]"
	}
}

// updateProgress creates a beautiful, creative timeline with wave patterns and animations
func (a *App) updateProgress() {
	if a.duration == 0 {
		// Animated no-song display
		beatPhase := int(time.Now().Unix()) % 4
		var animation string
		switch beatPhase {
		case 0: animation = "♪ ♫ ♪"
		case 1: animation = "♫ ♪ ♫"
		case 2: animation = "♪ ♫ ♪"
		case 3: animation = "♫ ♪ ♫"
		}
		a.progress.SetText(fmt.Sprintf("[magenta::b]════════════════════════ %s No song playing %s ════════════════════════[white]", animation, animation))
		return
	}
	
	progress := float64(a.position) / float64(a.duration)
	totalWidth := 80
	filled := int(progress * float64(totalWidth))
	
	// Create beautiful wave-like progress bar with dynamic colors
	var progressBar strings.Builder
	
	// Left border with music notes
	progressBar.WriteString("[cyan]♫[white]")
	
	// Build the wave pattern progress bar
	for i := 0; i < totalWidth; i++ {
		if i < filled {
			// Use consistent biggest size for filled portion
			var char string
			var color string
			
			// Dynamic colors based on position - all using biggest characters
			if i < totalWidth/4 {
				color = "[green::b]"
				char = "▇"  // Biggest character for this color
			} else if i < totalWidth/2 {
				color = "[yellow::b]"
				char = "█"  // Biggest character for this color
			} else if i < 3*totalWidth/4 {
				color = "[red::b]"
				char = "▇"  // Biggest character for this color
			} else {
				color = "[magenta::b]"
				char = "█"  // Biggest character for this color
			}
			progressBar.WriteString(color + char + "[white]")
	} else {
			// Empty portion with subtle pattern
			if i%3 == 0 {
				progressBar.WriteString("[gray]∙[white]")
			} else {
				progressBar.WriteString("[darkgray]·[white]")
			}
		}
	}
	
	// Right border with music notes
	progressBar.WriteString("[cyan]♫[white]")
	
	// Enhanced status with beautiful icons and animations
	var statusIcon, statusText, statusColor string
	if a.isPlaying {
		if a.isPaused {
			// Animated pause icon
			if time.Now().Unix()%2 == 0 {
				statusIcon = "⏸"
			} else {
				statusIcon = "⏸"
			}
			statusColor = "[yellow::b]"
			statusText = "PAUSED"
		} else {
			// Animated play icon
			playPhase := time.Now().UnixMilli() % 1000
			if playPhase < 500 {
				statusIcon = "▶"
			} else {
				statusIcon = "▷"
			}
			statusColor = "[green::b]"
			statusText = "PLAYING"
		}
	} else {
		statusIcon = "⏹"
		statusColor = "[red::b]"
		statusText = "STOPPED"
	}
	
	// Create beautiful time display with decorative elements
	currentTime := formatDuration(a.position)
	totalTime := formatDuration(a.duration)
	
	// Build the complete progress display
	progressText := fmt.Sprintf("%s [white]%s[cyan::b] %s [white]/ [cyan::b]%s [white]%s %s %s[white]", 
		progressBar.String(),
		statusColor,
		currentTime,
		totalTime,
		statusColor,
		statusIcon,
		statusText)
	
	a.progress.SetText(progressText)
}

// updateControls function removed - instructions moved to help window

// updateStatus updates the status bar
func (a *App) updateStatus() {
	status := fmt.Sprintf("[white]Songs: %d | %s | Score: %d | Press '/' to search, 'h' for help[white]", 
		len(a.songs), 
		a.getStatusText(),
		a.karaokeScore)
	
	a.statusBar.SetText(status)
}

// getStatusText returns the current status text
func (a *App) getStatusText() string {
	if a.isPlaying {
		if a.isPaused {
			return "PAUSED"
		}
		return "PLAYING"
	}
	return "STOPPED"
}

// getRepeatModeText returns the repeat mode display text
func (a *App) getRepeatModeText() string {
	if a.repeatMode {
		return "All"
	}
	return "Off"
}

// getShuffleModeText returns the shuffle mode display text
func (a *App) getShuffleModeText() string {
	if a.shuffleMode {
		return "On"
	}
	return "Off"
}

// showHelp shows help information
func (a *App) showHelp() {
	// Pause audio during help menu activation to prevent state issues
	wasPlaying := a.isPlaying && !a.isPaused
	if wasPlaying && a.player != nil {
		a.player.Pause()
	}
	
	// Create comprehensive help modal
	helpText := `[cyan]═══ BASIC CONTROLS ═══[white]                    [cyan]═══ ADVANCED FEATURES ═══[white]
[yellow]Space[white] - Play/Pause current song               [yellow]E[white] - Edit lyrics for current song
[yellow]s[white] - Stop playback and reset position          [yellow]F[white] - File management (move/rename/delete)
[yellow]↑/↓[white] - Navigate between songs                   [yellow]X[white] - Export data (performance/library)
[yellow]Enter[white] - Play the selected song                [yellow]J[white] - Jump to specific time (during playback)
[yellow]Tab[white] - Switch between search and song list     [yellow]I[white] - Show detailed song information
[yellow]/[white] - Focus on search box                       [yellow]K[white] - Toggle karaoke display mode
[yellow]ESC[white] - Clear search and return to song list    [yellow]C[white] - Clear all scores and start fresh

[cyan]═══ AUDIO CONTROLS ═══[white]                   [cyan]═══ QUICK ACCESS ═══[white]
[yellow]+/-[white] - Increase/Decrease volume               [yellow]1-9[white] - Jump to song by number (1-9)
[yellow]R[white] - Toggle repeat mode                       [yellow]0[white] - Jump to last song
[yellow]S[white] - Toggle shuffle mode                      [yellow]V[white] - Toggle mute/unmute
[yellow]←/→[white] - Seek backward/forward                   [yellow]M[white] - Mark song as favorite
[yellow]r[white] - Reload song library from files           [yellow]L[white] - Focus on lyrics panel

[cyan]═══ KARAOKE FEATURES ═══[white]
• [green]Real-time lyrics[white] highlight with the music • [green]Live scoring[white] system with accuracy tracking
• [green]Streak system[white] for consecutive hits        • [green]Audio visualizer[white] responds to music
• [green]Performance stats[white] and export capabilities • [green]5-line centered[white] lyrics display

[white]🎵 [yellow]Press [red]H[yellow], [red]Q[yellow], or [red]ESC[yellow] to close this help menu[white] 🎵`

	// Create a TextView for better control over sizing
	helpView := tview.NewTextView().
		SetText(helpText).
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetScrollable(true)
	
	helpView.SetBorder(true).
		SetTitle(" TUNEMINAL HELP ").
		SetTitleAlign(tview.AlignCenter)
	
	// Add input capture to the helpView as well
	helpView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			a.pages.RemovePage("help")
			a.app.SetFocus(a.songList)
			if wasPlaying && a.player != nil {
				a.player.Resume()
			}
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q', 'Q', 'h', 'H':
				a.pages.RemovePage("help")
				a.app.SetFocus(a.songList)
				if wasPlaying && a.player != nil {
					a.player.Resume()
				}
				return nil
			}
		}
		return event // Let other keys pass through
	})
	
	// Create a flexible layout to center the help view with proper sizing
	helpContainer := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 2, false).  // Top spacer
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).  // Left spacer
			AddItem(helpView, 0, 8, true).  // Help content (takes 8/10 of width)
			AddItem(nil, 0, 1, false),  // Right spacer
			0, 6, true).  // Help row (takes 6/10 of height)
		AddItem(nil, 0, 2, false)  // Bottom spacer
	
	// Add custom input capture for the help container
	helpContainer.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			a.pages.RemovePage("help")
			a.app.SetFocus(a.songList)
			if wasPlaying && a.player != nil {
				a.player.Resume()
			}
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q', 'Q', 'h', 'H':
				a.pages.RemovePage("help")
				a.app.SetFocus(a.songList)
				if wasPlaying && a.player != nil {
					a.player.Resume()
				}
				return nil
			}
		}
		return event // Let other keys pass through
	})
	
	a.pages.AddPage("help", helpContainer, true, true)
	a.app.SetFocus(helpView) // Focus on the helpView for better key capture
}

// Navigation functions
func (a *App) navigateUp() {
	if a.currentSong > 0 {
		a.currentSong--
		a.updateSongList()
		a.updateNowPlaying()
		a.updateKaraokeLyrics()
		// Ensure focus stays on song list
		a.app.SetFocus(a.songList)
	}
}

func (a *App) navigateDown() {
	if a.currentSong < len(a.songs)-1 {
		a.currentSong++
		a.updateSongList()
	a.updateNowPlaying()
	a.updateKaraokeLyrics()
		// Ensure focus stays on song list
		a.app.SetFocus(a.songList)
	}
}


func (a *App) playSelectedSong() {
	// Prevent multiple simultaneous play attempts
	if a.isLoading {
		return
	}
	
	if a.currentSong >= 0 && a.currentSong < len(a.songs) {
		// Get the selected song index from the song list
		selectedIndex := a.songList.GetCurrentItem()
		
		// If pressing Enter on the same currently playing song, toggle play/pause
		if selectedIndex == a.currentSong && a.isPlaying {
			a.togglePlayPause()
		} else {
			// Different song or not playing, start new playback
			a.currentSong = selectedIndex
			a.play()
		}
	}
}

func (a *App) onSearchChanged(text string) {
	// Filter songs based on search text
	a.filterAndUpdateSongList(text)
}

// filterAndUpdateSongList filters songs based on search text and updates the display
func (a *App) filterAndUpdateSongList(searchText string) {
	a.songList.Clear()
	
	// If no search text, show all songs
	if searchText == "" {
		for i, song := range a.songs {
			// Format: "Title - Artist [Duration]"
			mainText := fmt.Sprintf("%s - %s", song.Title, song.Artist)
			secondaryText := fmt.Sprintf("[%02d:%02d]", 
				int(song.Duration.Minutes()), 
				int(song.Duration.Seconds())%60)
			
			a.songList.AddItem(mainText, secondaryText, 0, func() {
				a.currentSong = i
				a.playSelectedSong()
			})
		}
		return
	}
	
	// Filter songs that match search text (case insensitive)
	searchLower := strings.ToLower(searchText)
	matchedIndices := []int{}
	
	for i, song := range a.songs {
		titleMatch := strings.Contains(strings.ToLower(song.Title), searchLower)
		artistMatch := strings.Contains(strings.ToLower(song.Artist), searchLower)
		
		if titleMatch || artistMatch {
			matchedIndices = append(matchedIndices, i)
			
			// Format: "Title - Artist [Duration]" with search highlighting
			mainText := fmt.Sprintf("%s - %s", song.Title, song.Artist)
			secondaryText := fmt.Sprintf("[%02d:%02d] [green]✓[white]", 
				int(song.Duration.Minutes()), 
				int(song.Duration.Seconds())%60)
			
			a.songList.AddItem(mainText, secondaryText, 0, func(index int) func() {
				return func() {
					a.currentSong = index
					a.playSelectedSong()
				}
			}(i))
		}
	}
	
	// Update status to show search results
	if len(matchedIndices) == 0 {
		a.songList.AddItem("[red]No songs found[white]", 
			fmt.Sprintf("No matches for '%s'", searchText), 0, nil)
	}
}

// Playback controls
func (a *App) play() {
	if a.currentSong < 0 || a.currentSong >= len(a.songs) {
		return
	}

	// Set loading flag to prevent multiple simultaneous attempts
	a.isLoading = true
	defer func() {
		a.isLoading = false
	}()

	song := a.songs[a.currentSong]

	// Load lyrics for this song
	if song.LyricsPath != "" {
		a.loadLyricsFromFile(song.LyricsPath)
	} else {
		a.lyricLines = []LyricLine{
			{Time: 0 * time.Second, Text: "No lyrics available", Index: 0, IsActive: false, IsHit: false},
		}
	}

	// Reset karaoke state only for NEW playback (not resume)
	if !a.isPaused {
		a.karaokeScore = 0
		a.streak = 0
		a.accuracy = 0.0
		a.totalLyrics = len(a.lyricLines)
		a.hitLyrics = 0
		for i := range a.lyricLines {
			a.lyricLines[i].IsHit = false
			a.lyricLines[i].IsActive = false
		}
	}

	// Real audio playback with optimized responsiveness
	if a.player != nil {
		// Load the audio file (this is cached after first load)
		if err := a.player.LoadFile(song.Path); err != nil {
			a.handleError(err, "Load Audio File")
			return
		}

		// Apply current volume setting
		a.player.SetVolume(a.volume)

		// If resuming from pause, seek to current position
		if a.isPaused && a.position > 0 {
			if err := a.player.SeekTo(a.position); err != nil {
				a.handleError(err, "Seek to Position")
				// Continue anyway, will start from beginning
			}
		}

		// Start playback immediately - don't wait for UI updates
		if err := a.player.Play(); err != nil {
			a.handleError(err, "Start Playback")
			return
		}

		// Set UI state (after audio starts)
		a.isPlaying = true
		a.isPaused = false
		a.duration = song.Duration

		// Update UI in background to not block audio
		go func() {
			a.app.QueueUpdateDraw(func() {
				a.updateAllDisplays()
			})
		}()

		// Start position tracking for UI updates
		go a.trackRealPlayback()
	}
}

func (a *App) togglePlayPause() {
	if a.isPlaying && !a.isPaused {
		// Currently playing, so pause
		a.pause()
	} else {
		// Currently paused or stopped, so start/resume
		if a.currentSong >= 0 {
			a.play()
		}
	}
}

func (a *App) pause() {
	if a.player != nil {
		a.player.Pause()
	}
	a.isPaused = true
	a.isPlaying = false
	a.updateAllDisplays()
}

// trackRealPlayback tracks real audio playback position
func (a *App) trackRealPlayback() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if !a.isPlaying {
			break
		}

		// Get real position from audio player
		if a.player != nil {
			a.position = a.player.GetPosition()
		}

		// Check if song is finished
		if !a.player.IsPlaying() || a.position >= a.duration {
			a.position = a.duration
			a.isPlaying = false
			a.isPaused = false
			// Ensure focus returns to song list when song ends
			a.app.QueueUpdateDraw(func() {
				a.app.SetFocus(a.songList)
				a.updateAllDisplays()
			})
			break
		}

		a.app.QueueUpdateDraw(func() {
			a.updateNowPlaying()
			a.updateProgress()
			a.updateKaraokeLyrics()
			a.updateVisualizer()
			a.updateScore()
			a.updateSongList()
		})
	}
}

func (a *App) stop() {
	// Ensure we stop cleanly to prevent corruption
	if a.player != nil {
		a.player.Stop()
	}

	// Reset all playback state
	a.isPlaying = false
	a.isPaused = false
	a.position = 0

	// Reset loading flag in case it was set
	a.isLoading = false

	// Reset scoring and visualizer state to prevent glitches
	a.karaokeScore = 0
	a.streak = 0
	a.accuracy = 0.0
	a.hitLyrics = 0
	a.totalLyrics = 0

	// Reset visualizer bars
	for i := range a.visualizerBars {
		a.visualizerBars[i] = 0
	}
	a.beatPhase = 0

	// Reset lyric states
	for i := range a.lyricLines {
		a.lyricLines[i].IsHit = false
		a.lyricLines[i].IsActive = false
	}

	a.updateAllDisplays()
	// Ensure focus returns to song list
	a.app.SetFocus(a.songList)
}

func (a *App) next() {
	if len(a.songs) == 0 {
		return
	}
	
	a.currentSong = (a.currentSong + 1) % len(a.songs)
	a.updateSongList()
	a.play()
}

func (a *App) previous() {
	if len(a.songs) == 0 {
		return
	}
	
	a.currentSong = (a.currentSong - 1 + len(a.songs)) % len(a.songs)
	a.updateSongList()
	a.play()
}

// Volume control functions
func (a *App) increaseVolume() {
	if a.volume < 1.0 {
		a.volume = a.volume + 0.1
		if a.volume > 1.0 {
			a.volume = 1.0
		}
		if a.player != nil {
			a.player.SetVolume(a.volume)
		}
		a.updateNowPlaying()
		a.saveConfig()
	}
}

func (a *App) decreaseVolume() {
	if a.volume > 0.0 {
		a.volume = a.volume - 0.1
		if a.volume < 0.0 {
			a.volume = 0.0
		}
		if a.player != nil {
			a.player.SetVolume(a.volume)
		}
		a.updateNowPlaying()
		a.saveConfig()
	}
}

func (a *App) toggleRepeat() {
	a.repeatMode = !a.repeatMode
	a.updateNowPlaying()
	a.saveConfig()
}

func (a *App) toggleShuffle() {
	a.shuffleMode = !a.shuffleMode
	a.updateNowPlaying()
	a.saveConfig()
}

// saveConfig saves the current configuration to file
func (a *App) saveConfig() {
	if a.appConfig != nil {
		a.appConfig.DefaultVolume = a.volume
		a.appConfig.ShuffleMode = a.shuffleMode
		a.appConfig.RepeatMode = a.repeatMode
		a.appConfig.SaveConfig(config.GetConfigPath())
	}
}

// Playlist management functions
func (a *App) createPlaylist(name, description string) error {
	_, err := a.playlistManager.CreatePlaylist(name, description)
	return err
}

func (a *App) loadPlaylist(playlistName string) error {
	songPaths, err := a.playlistManager.GetPlaylistSongs(playlistName)
	if err != nil {
		return err
	}

	// Clear current songs
	a.songs = []Song{}
	a.currentSong = -1

	// Load songs from playlist
	for _, path := range songPaths {
		// Check if file exists
		if _, err := os.Stat(path); err == nil {
			// Try to get metadata
			meta, err := metadata.GetRealMetadata(path)
			if err == nil {
				song := Song{
					Title:      meta.Title,
					Artist:     meta.Artist,
					Path:       meta.Path,
					LyricsPath: a.findLyricsFile(meta.Path),
					Duration:   meta.Duration,
				}
				a.songs = append(a.songs, song)
			}
		}
	}

	a.currentPlaylist = playlistName

	// Update displays
	a.updateAllDisplays()

	return nil
}

func (a *App) addSongToPlaylist(playlistName string) error {
	if a.currentSong >= 0 && a.currentSong < len(a.songs) {
		songPath := a.songs[a.currentSong].Path
		return a.playlistManager.AddSongToPlaylist(playlistName, songPath)
	}
	return fmt.Errorf("no song selected")
}

func (a *App) getPlaylistList() []string {
	playlists, err := a.playlistManager.ListPlaylists()
	if err != nil {
		return []string{}
	}
	return playlists
}

// Lyrics Editor functions
func (a *App) openLyricsEditor() {
	if a.currentSong < 0 || a.currentSong >= len(a.songs) {
		return
	}

	song := a.songs[a.currentSong]

	// Load existing lyrics if available
	if song.LyricsPath != "" {
		if err := a.lyricsEditor.LoadLyricsFromFile(song.LyricsPath); err != nil {
			// Start with empty editor if loading fails
			a.lyricsEditor = lyrics.NewLyricEditor()
		}
	} else {
		// Start with empty editor for new lyrics
		a.lyricsEditor = lyrics.NewLyricEditor()
	}

	// Create lyrics editor modal
	a.showLyricsEditor(song)
}

func (a *App) saveLyrics() {
	if a.currentSong < 0 || a.currentSong >= len(a.songs) {
		return
	}

	song := a.songs[a.currentSong]

	// Generate lyrics file path
	lyricsPath := a.findLyricsFile(song.Path)
	if lyricsPath == "" {
		// Create new lyrics file path
		ext := filepath.Ext(song.Path)
		lyricsPath = strings.TrimSuffix(song.Path, ext) + ".lrc"
	}

	// Save lyrics
	if err := a.lyricsEditor.SaveLyricsToFile(lyricsPath); err != nil {
		a.handleError(err, "Lyrics Save")
		return
	}

	// Update song's lyrics path
	song.LyricsPath = lyricsPath
	a.songs[a.currentSong] = song

	// Reload lyrics in main display
	a.loadLyricsFromFile(lyricsPath)
}

// showLyricsEditor displays the lyrics editor modal
func (a *App) showLyricsEditor(song Song) {
	lyricsLines := a.lyricsEditor.GetLyricsLines()

	// Convert lyrics lines to interface{} for display
	displayLines := make([]interface{}, len(lyricsLines))
	for i, line := range lyricsLines {
		displayLines[i] = map[string]interface{}{
			"time": line.Time,
			"text": line.Text,
		}
	}

	// Create a text view for the lyrics editor
	editorText := a.createLyricsEditorContent(song, displayLines)

	lyricsEditorModal := tview.NewModal().
		SetText(editorText).
		AddButtons([]string{"Save", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Save" {
				a.saveLyrics()
			}
			a.pages.RemovePage("lyrics-editor")
			a.app.SetFocus(a.songList)
		})

	// Set modal title
	lyricsEditorModal.SetTitle("Lyrics Editor - " + song.Title)

	a.pages.AddPage("lyrics-editor", lyricsEditorModal, true, true)
	a.app.SetFocus(lyricsEditorModal)
}

// createLyricsEditorContent creates the content for the lyrics editor
func (a *App) createLyricsEditorContent(song Song, lyricsLines []interface{}) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("[yellow]Editing lyrics for: %s - %s[white]\n\n", song.Title, song.Artist))

	if len(lyricsLines) == 0 {
		content.WriteString("[cyan]No lyrics loaded. Start by adding your first lyric line.[white]\n\n")
		content.WriteString("[white]Format: [mm:ss.xx] Your lyrics here[white]\n")
		content.WriteString("[white]Example: [00:30.50] Welcome to the song![white]\n\n")
	} else {
		content.WriteString("[cyan]Current Lyrics:[white]\n")
		for i, lineInterface := range lyricsLines {
			// Convert interface{} to map for display
			if lineMap, ok := lineInterface.(map[string]interface{}); ok {
				timeInterface, hasTime := lineMap["time"]
				textInterface, hasText := lineMap["text"]

				if hasTime && hasText {
					// Format time as string for display
					timeStr := "00:00.00" // Placeholder - would need actual time formatting
					if timeDuration, ok := timeInterface.(time.Duration); ok {
						minutes := int(timeDuration.Minutes())
						seconds := int(timeDuration.Seconds()) % 60
						centiseconds := int(timeDuration.Milliseconds()) % 1000 / 10
						timeStr = fmt.Sprintf("[%02d:%02d.%02d]", minutes, seconds, centiseconds)
					}

					text := ""
					if textStr, ok := textInterface.(string); ok {
						text = textStr
					}

					content.WriteString(fmt.Sprintf("[yellow]%d.[white] %s %s\n", i+1, timeStr, text))
				}
			}
		}
		content.WriteString("\n")
	}

	content.WriteString("[green]Instructions:[white]\n")
	content.WriteString("• Edit the lyrics above with proper timing\n")
	content.WriteString("• Use format [mm:ss.xx] for timing\n")
	content.WriteString("• Press [yellow]Save[white] to save changes\n")
	content.WriteString("• Press [yellow]Cancel[white] to discard changes\n")

	return content.String()
}

// File Management functions
func (a *App) moveSongToDirectory(song Song, newDir string) error {
	if a.currentSong < 0 || a.currentSong >= len(a.songs) {
		return fmt.Errorf("no song selected")
	}

	// Create new path
	newPath := filepath.Join(newDir, filepath.Base(song.Path))

	// Check if destination file already exists
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("file already exists at destination")
	}

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Move the file
	if err := os.Rename(song.Path, newPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	// Update song path
	a.songs[a.currentSong].Path = newPath

	// Update lyrics path if it exists
	if song.LyricsPath != "" {
		lyricsFileName := filepath.Base(song.LyricsPath)
		newLyricsPath := filepath.Join(newDir, lyricsFileName)
		if _, err := os.Stat(song.LyricsPath); err == nil {
			os.Rename(song.LyricsPath, newLyricsPath)
			a.songs[a.currentSong].LyricsPath = newLyricsPath
		}
	}

	return nil
}

func (a *App) renameSong(song Song, newName string) error {
	if a.currentSong < 0 || a.currentSong >= len(a.songs) {
		return fmt.Errorf("no song selected")
	}

	// Create new path with new filename
	dir := filepath.Dir(song.Path)
	ext := filepath.Ext(song.Path)
	newPath := filepath.Join(dir, newName+ext)

	// Check if destination file already exists
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("file with that name already exists")
	}

	// Rename the file
	if err := os.Rename(song.Path, newPath); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	// Update song path
	a.songs[a.currentSong].Path = newPath

	// Update lyrics path if it exists
	if song.LyricsPath != "" {
		lyricsDir := filepath.Dir(song.LyricsPath)
		lyricsExt := filepath.Ext(song.LyricsPath)
		newLyricsPath := filepath.Join(lyricsDir, newName+lyricsExt)
		if _, err := os.Stat(song.LyricsPath); err == nil {
			os.Rename(song.LyricsPath, newLyricsPath)
			a.songs[a.currentSong].LyricsPath = newLyricsPath
		}
	}

	return nil
}

func (a *App) deleteSong(song Song) error {
	if a.currentSong < 0 || a.currentSong >= len(a.songs) {
		return fmt.Errorf("no song selected")
	}

	// Delete audio file
	if err := os.Remove(song.Path); err != nil {
		return fmt.Errorf("failed to delete audio file: %w", err)
	}

	// Delete lyrics file if it exists
	if song.LyricsPath != "" {
		if _, err := os.Stat(song.LyricsPath); err == nil {
			os.Remove(song.LyricsPath)
		}
	}

	// Remove song from library
	a.songs = append(a.songs[:a.currentSong], a.songs[a.currentSong+1:]...)

	// Adjust current song index
	if a.currentSong >= len(a.songs) {
		a.currentSong = len(a.songs) - 1
	}

	return nil
}

// showFileManager displays the file management modal
func (a *App) showFileManager() {
	if a.currentSong < 0 || a.currentSong >= len(a.songs) {
		return
	}

	song := a.songs[a.currentSong]

	fileManagerModal := tview.NewModal().
		SetText(a.createFileManagerContent(song)).
		AddButtons([]string{"Move", "Rename", "Delete", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			switch buttonLabel {
			case "Move":
				a.showMoveDialog(song)
			case "Rename":
				a.showRenameDialog(song)
			case "Delete":
				a.showDeleteConfirmation(song)
			}
			a.pages.RemovePage("file-manager")
			a.app.SetFocus(a.songList)
		})

	fileManagerModal.SetTitle("File Manager - " + song.Title)
	a.pages.AddPage("file-manager", fileManagerModal, true, true)
	a.app.SetFocus(fileManagerModal)
}

// createFileManagerContent creates the content for the file manager
func (a *App) createFileManagerContent(song Song) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("[yellow]Managing file: %s[white]\n", song.Title))
	content.WriteString(fmt.Sprintf("[white]Artist: %s[white]\n", song.Artist))
	content.WriteString(fmt.Sprintf("[white]Current path: %s[white]\n", song.Path))
	content.WriteString(fmt.Sprintf("[white]Duration: %s[white]\n", formatDuration(song.Duration)))

	if song.LyricsPath != "" {
		content.WriteString(fmt.Sprintf("[white]Lyrics: %s[white]\n", song.LyricsPath))
	} else {
		content.WriteString("[white]Lyrics: [red]Not available[white]\n")
	}

	content.WriteString("\n[green]Choose an action:[white]\n")
	content.WriteString("[yellow]Move[white] - Move file to different directory\n")
	content.WriteString("[yellow]Rename[white] - Rename the file\n")
	content.WriteString("[yellow]Delete[white] - Delete the file permanently\n")
	content.WriteString("[yellow]Cancel[white] - Return to music library\n")

	return content.String()
}

// showMoveDialog shows a dialog for moving files
func (a *App) showMoveDialog(song Song) {
	directoryInput := tview.NewInputField().SetLabel("Destination Directory").SetText("").SetFieldWidth(50)

	form := tview.NewForm().
		AddFormItem(directoryInput).
		AddButton("Move", func() {
			directory := directoryInput.GetText()
			if directory != "" {
				if err := a.moveSongToDirectory(song, directory); err != nil {
					a.handleError(err, "Move File")
				} else {
					a.updateAllDisplays()
					a.showMessage("✅ File moved successfully!")
				}
			} else {
				a.showWarning("Please enter a destination directory")
			}
			a.pages.RemovePage("move-dialog")
			a.app.SetFocus(a.songList)
		}).
		AddButton("Cancel", func() {
			a.pages.RemovePage("move-dialog")
			a.app.SetFocus(a.songList)
		})

	form.SetTitle("Move File").SetBorder(true)
	a.pages.AddPage("move-dialog", form, true, true)
}

// showRenameDialog shows a dialog for renaming files
func (a *App) showRenameDialog(song Song) {
	// Get current filename without extension
	currentName := strings.TrimSuffix(filepath.Base(song.Path), filepath.Ext(song.Path))

	newNameInput := tview.NewInputField().SetLabel("New Name").SetText(currentName).SetFieldWidth(30)

	form := tview.NewForm().
		AddFormItem(newNameInput).
		AddButton("Rename", func() {
			newName := newNameInput.GetText()
			if newName == "" {
				a.showWarning("Please enter a new name")
				return
			}
			if newName == currentName {
				a.showWarning("New name is the same as current name")
				return
			}
			if err := a.renameSong(song, newName); err != nil {
				a.handleError(err, "Rename File")
			} else {
				a.updateAllDisplays()
				a.showMessage("✅ File renamed successfully!")
			}
			a.pages.RemovePage("rename-dialog")
			a.app.SetFocus(a.songList)
		}).
		AddButton("Cancel", func() {
			a.pages.RemovePage("rename-dialog")
			a.app.SetFocus(a.songList)
		})

	form.SetTitle("Rename File").SetBorder(true)
	a.pages.AddPage("rename-dialog", form, true, true)
}

// showDeleteConfirmation shows a confirmation dialog for deleting files
func (a *App) showDeleteConfirmation(song Song) {
	confirmModal := tview.NewModal().
		SetText(fmt.Sprintf("[red]Are you sure you want to delete:[white]\n\n%s - %s\n\n[red]This action cannot be undone![white]\n\n[dim]Press 'y' to confirm, 'n' to cancel, or Tab+Enter for buttons[white]", song.Title, song.Artist)).
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Delete" {
				if err := a.deleteSong(song); err != nil {
					a.handleError(err, "Delete File")
				} else {
					a.updateAllDisplays()
					a.showMessage("🗑️ File deleted successfully!")
				}
			}
			a.pages.RemovePage("delete-confirm")
			a.app.SetFocus(a.songList)
		})

	// Add keyboard shortcuts for quick actions (non-conflicting keys)
	confirmModal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'y', 'Y':
				// 'y' confirms deletion
				if err := a.deleteSong(song); err != nil {
					a.handleError(err, "Delete File")
				} else {
					a.updateAllDisplays()
					a.showMessage("🗑️ File deleted successfully!")
				}
				a.pages.RemovePage("delete-confirm")
				a.app.SetFocus(a.songList)
				return nil
			case 'n', 'N':
				// 'n' cancels deletion
				a.pages.RemovePage("delete-confirm")
				a.app.SetFocus(a.songList)
				return nil
			}
		}
		return event
	})

	confirmModal.SetTitle("Delete Confirmation")
	a.pages.AddPage("delete-confirm", confirmModal, true, true)
	a.app.SetFocus(confirmModal)
}

// Export/Import functions
func (a *App) exportPerformanceData(format string) error {
	// Create performance data from current session
	performanceData := []export.PerformanceData{}

	// Add current performance if playing
	if a.currentSong >= 0 && a.currentSong < len(a.songs) {
		song := a.songs[a.currentSong]
		perf := export.PerformanceData{
			Date:      time.Now(),
			SongTitle: song.Title,
			Artist:    song.Artist,
			Score:     a.karaokeScore,
			Streak:    a.streak,
			Accuracy:  a.accuracy,
			Duration:  formatDuration(song.Duration),
		}
		performanceData = append(performanceData, perf)
	}

	return a.exportManager.ExportPerformanceData(performanceData, format)
}

func (a *App) exportLibraryData(format string) error {
	// Convert songs to library data format
	libraryData := make([]export.LibraryData, len(a.songs))
	for i, song := range a.songs {
		libraryData[i] = export.LibraryData{
			Title:      song.Title,
			Artist:     song.Artist,
			Path:       song.Path,
			LyricsPath: song.LyricsPath,
			Duration:   formatDuration(song.Duration),
			Format:     "mp3", // Could be enhanced to detect actual format
			Size:       0,     // Would need to get file size
		}
	}

	return a.exportManager.ExportLibraryData(libraryData, format)
}

func (a *App) showExportDialog() {
	exportModal := tview.NewModal().
		SetText(a.createExportDialogContent()).
		AddButtons([]string{"Performance JSON", "Performance CSV", "Library JSON", "Library CSV", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			switch buttonLabel {
			case "Performance JSON":
				if err := a.exportPerformanceData("json"); err != nil {
					a.handleError(err, "Performance JSON Export")
				} else {
					a.showExportSuccess("Performance data exported as JSON")
				}
			case "Performance CSV":
				if err := a.exportPerformanceData("csv"); err != nil {
					a.handleError(err, "Performance CSV Export")
				} else {
					a.showExportSuccess("Performance data exported as CSV")
				}
			case "Library JSON":
				if err := a.exportLibraryData("json"); err != nil {
					a.handleError(err, "Library JSON Export")
				} else {
					a.showExportSuccess("Library data exported as JSON")
				}
			case "Library CSV":
				if err := a.exportLibraryData("csv"); err != nil {
					a.handleError(err, "Library CSV Export")
				} else {
					a.showExportSuccess("Library data exported as CSV")
				}
			}
			a.pages.RemovePage("export-dialog")
			a.app.SetFocus(a.songList)
		})

	exportModal.SetTitle("Export Data")
	a.pages.AddPage("export-dialog", exportModal, true, true)
	a.app.SetFocus(exportModal)
}

func (a *App) createExportDialogContent() string {
	var content strings.Builder

	content.WriteString("[yellow]Export Options:[white]\n\n")

	content.WriteString("[cyan]Performance Data:[white]\n")
	content.WriteString("• [yellow]Performance JSON[white] - Export karaoke performance statistics as JSON\n")
	content.WriteString("• [yellow]Performance CSV[white] - Export karaoke performance statistics as CSV\n\n")

	content.WriteString("[cyan]Library Data:[white]\n")
	content.WriteString("• [yellow]Library JSON[white] - Export music library information as JSON\n")
	content.WriteString("• [yellow]Library CSV[white] - Export music library information as CSV\n\n")

	content.WriteString("[green]Files will be saved to:[white]\n")
	content.WriteString(fmt.Sprintf("%s\n\n", a.exportManager.GetExportPath()))

	content.WriteString("[white]Press the export option you want to use.[white]")

	return content.String()
}

func (a *App) showExportSuccess(message string) {
	successModal := tview.NewModal().
		SetText(fmt.Sprintf("[green]✅ %s[white]\n\nFiles saved to: %s\n\n[dim]Press 'o' for OK or Tab+Enter to use button[white]", message, a.exportManager.GetExportPath())).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("export-success")
			a.app.SetFocus(a.songList)
		})

	// Add keyboard shortcuts for closing (non-conflicting keys)
	successModal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'o', 'O':
				// 'o' for OK
				a.pages.RemovePage("export-success")
				a.app.SetFocus(a.songList)
				return nil
			}
		}
		return event
	})

	a.pages.AddPage("export-success", successModal, true, true)
	a.app.SetFocus(successModal)
}

// showMessage displays a temporary message to the user
func (a *App) showMessage(message string) {
	messageModal := tview.NewModal().
		SetText(message + "\n\n[dim]Press 'o' for OK or Tab+Enter to use button[white]").
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("message")
			a.app.SetFocus(a.songList)
		})

	// Add keyboard shortcuts for closing (non-conflicting keys)
	messageModal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'o', 'O':
				// 'o' for OK
				a.pages.RemovePage("message")
				a.app.SetFocus(a.songList)
				return nil
			}
		}
		return event
	})

	a.pages.AddPage("message", messageModal, true, true)
	a.app.SetFocus(messageModal)
}

// showError displays an error message to the user
func (a *App) showError(message string) {
	errorModal := tview.NewModal().
		SetText("[red]❌ Error:[white]\n\n" + message + "\n\n[dim]Press 'o' for OK or Tab+Enter to use button[white]").
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("error")
			a.app.SetFocus(a.songList)
		})

	// Add keyboard shortcuts for closing (non-conflicting keys)
	errorModal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'o', 'O':
				// 'o' for OK
				a.pages.RemovePage("error")
				a.app.SetFocus(a.songList)
				return nil
			}
		}
		return event
	})

	a.pages.AddPage("error", errorModal, true, true)
	a.app.SetFocus(errorModal)
}

// showWarning displays a warning message to the user
func (a *App) showWarning(message string) {
	warningModal := tview.NewModal().
		SetText("[yellow]⚠️ Warning:[white]\n\n" + message + "\n\n[dim]Press 'o' for OK or Tab+Enter to use button[white]").
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("warning")
			a.app.SetFocus(a.songList)
		})

	// Add keyboard shortcuts for closing (non-conflicting keys)
	warningModal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'o', 'O':
				// 'o' for OK
				a.pages.RemovePage("warning")
				a.app.SetFocus(a.songList)
				return nil
			}
		}
		return event
	})

	a.pages.AddPage("warning", warningModal, true, true)
	a.app.SetFocus(warningModal)
}

// handleError provides centralized error handling with user feedback
func (a *App) handleError(err error, context string) {
	if err == nil {
		return
	}

	errorMsg := fmt.Sprintf("Context: %s\nError: %s", context, err.Error())
	a.showError(errorMsg)

	// Log error to status bar for debugging
	a.statusBar.SetText(fmt.Sprintf("[red]Error in %s: %s[white]", context, err.Error()))

	// Clear error message after 5 seconds
	go func() {
		time.Sleep(5 * time.Second)
		a.app.QueueUpdateDraw(func() {
			// Only clear if it's still the same error message
			if strings.Contains(a.statusBar.GetText(false), context) {
				a.updateStatus()
			}
		})
	}()
}

// showJumpToTimeDialog shows a dialog for jumping to a specific time
func (a *App) showJumpToTimeDialog() {
	timeInput := tview.NewInputField().SetLabel("Jump to time (mm:ss)").SetText("").SetFieldWidth(10)

	form := tview.NewForm().
		AddFormItem(timeInput).
		AddButton("Jump", func() {
			timeStr := timeInput.GetText()
			if timeStr == "" {
				a.showWarning("Please enter a time")
				return
			}

			// Parse time string and jump to position
			duration, err := parseTimeString(timeStr)
			if err != nil {
				a.showError("Invalid time format. Use mm:ss (e.g., 01:30)")
				return
			}

			// Check if time is within song duration
			if a.currentSong >= 0 && a.currentSong < len(a.songs) {
				songDuration := a.songs[a.currentSong].Duration
				if duration > songDuration {
					a.showWarning(fmt.Sprintf("Time exceeds song duration (%s)", formatDuration(songDuration)))
					return
				}
			}

			if a.player != nil {
				if err := a.player.SeekTo(duration); err != nil {
					a.handleError(err, "Seek to Time")
					return
				}
				a.position = duration
				a.updateAllDisplays()
				a.showMessage(fmt.Sprintf("✅ Jumped to %s", timeStr))
			} else {
				a.showError("No audio player available")
				return
			}

			a.pages.RemovePage("jump-dialog")
			a.app.SetFocus(a.songList)
		}).
		AddButton("Cancel", func() {
			a.pages.RemovePage("jump-dialog")
			a.app.SetFocus(a.songList)
		})

	form.SetTitle("Jump to Time").SetBorder(true)
	a.pages.AddPage("jump-dialog", form, true, true)
}

// parseTimeString parses a time string in mm:ss format
func parseTimeString(timeStr string) (time.Duration, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid format")
	}

	minutes, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}

	seconds, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	return time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second, nil
}

// showSongInfo displays detailed information about the current song
func (a *App) showSongInfo() {
	if a.currentSong < 0 || a.currentSong >= len(a.songs) {
		return
	}

	song := a.songs[a.currentSong]

	var info strings.Builder
	info.WriteString(fmt.Sprintf("[yellow]Song Information:[white]\n\n"))
	info.WriteString(fmt.Sprintf("[cyan]Title:[white] %s\n", song.Title))
	info.WriteString(fmt.Sprintf("[cyan]Artist:[white] %s\n", song.Artist))
	info.WriteString(fmt.Sprintf("[cyan]Duration:[white] %s\n", formatDuration(song.Duration)))
	info.WriteString(fmt.Sprintf("[cyan]File Path:[white] %s\n", song.Path))

	if song.LyricsPath != "" {
		info.WriteString(fmt.Sprintf("[cyan]Lyrics:[white] %s\n", song.LyricsPath))
	} else {
		info.WriteString("[cyan]Lyrics:[white] [red]Not available[white]\n")
	}

	info.WriteString("\n[green]Press any key to close[white]")

	infoModal := tview.NewModal().
		SetText(info.String()).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("song-info")
			a.app.SetFocus(a.songList)
		})

	a.pages.AddPage("song-info", infoModal, true, true)
}

// toggleKaraokeDisplay toggles the visibility of karaoke lyrics
func (a *App) toggleKaraokeDisplay() {
	// For now, just show a message - could be extended to actually hide/show lyrics
	a.showMessage("🎤 Karaoke display toggled!")
}

// Seek functions
func (a *App) seekForward() {
	if a.currentSong >= 0 && a.currentSong < len(a.songs) && a.isPlaying {
		seekStep := time.Duration(a.appConfig.SeekStep) * time.Second
		newPosition := a.position + seekStep
		if newPosition > a.duration {
			newPosition = a.duration
		}

		if a.player != nil {
			if err := a.player.SeekTo(newPosition); err == nil {
				a.position = newPosition
				a.updateAllDisplays()
			}
		}
	}
}

func (a *App) seekBackward() {
	if a.currentSong >= 0 && a.currentSong < len(a.songs) && a.isPlaying {
		seekStep := time.Duration(a.appConfig.SeekStep) * time.Second
		newPosition := a.position - seekStep
		if newPosition < 0 {
			newPosition = 0
		}

		if a.player != nil {
			if err := a.player.SeekTo(newPosition); err == nil {
				a.position = newPosition
				a.updateAllDisplays()
			}
		}
	}
}

func (a *App) quit() {
	if a.player != nil {
		a.player.Stop()
	}
	a.app.Stop()
}

// Helper functions
func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// Run starts the application
func (a *App) Run() error {
	return a.app.Run()
}


func main() {
	// Add crash recovery
	defer func() {
		if r := recover(); r != nil {
			// Silent recovery
		}
	}()
	
	// Create and run app
	app := NewApp()
	
	if err := app.Run(); err != nil {
		// Silent exit
	}
}
