package player

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/wav"
	"github.com/faiface/beep"
)

// AudioPlayer handles audio playback using stable Oto library
type AudioPlayer struct {
	otoContext   *oto.Context
	player       *oto.Player
	mutex        sync.RWMutex
	isLoaded     bool
	isPlaying    bool
	isPaused     bool
	currentFile  string
	audioData    []byte
	sampleRate   int
	channels     int
	duration     time.Duration
	position     time.Duration
	playbackDone chan struct{}
	volume       float64 // Volume level from 0.0 to 1.0
}

// LyricEntry represents a single lyric entry with timing
type LyricEntry struct {
	Time time.Duration
	Text string
}

// NewAudioPlayer creates a new audio player using Oto
func NewAudioPlayer() *AudioPlayer {
	return &AudioPlayer{
		playbackDone: make(chan struct{}),
		sampleRate:   44100,
		channels:     2,
		volume:       1.0, // Default volume (100%)
	}
}

// initializeOto initializes the Oto context if not already done
func (p *AudioPlayer) initializeOto() error {
	if p.otoContext != nil {
		return nil
	}

	// Initialize Oto context with optimized buffer size for low latency
	op := &oto.NewContextOptions{
		SampleRate:   p.sampleRate,
		ChannelCount: p.channels,
		Format:       oto.FormatSignedInt16LE,
		BufferSize:   1024, // Smaller buffer for lower latency (was default ~4096)
	}

	ctx, readyChan, err := oto.NewContext(op)
	if err != nil {
		return fmt.Errorf("failed to create Oto context: %w", err)
	}

	// Wait for the context to be ready
	<-readyChan

	p.otoContext = ctx
	return nil
}

// LoadFile loads an audio file using Oto for stable playback
func (p *AudioPlayer) LoadFile(filename string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Stop any current playback
	p.stopInternal()

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("audio file not found: %s", filename)
	}

	// Open the audio file
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Determine file type and decode
	ext := strings.ToLower(filepath.Ext(filename))
	var streamer beep.StreamSeekCloser
	var format beep.Format

	switch ext {
	case ".mp3":
		streamer, format, err = mp3.Decode(file)
		if err != nil {
			return fmt.Errorf("failed to decode MP3: %w", err)
		}
	case ".wav":
		streamer, format, err = wav.Decode(file)
		if err != nil {
			return fmt.Errorf("failed to decode WAV: %w", err)
		}
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}
	defer streamer.Close()

	// Set audio parameters from the decoded format
	p.sampleRate = int(format.SampleRate)
	p.channels = format.NumChannels

	// Initialize Oto with the correct format
	if err := p.initializeOto(); err != nil {
		return fmt.Errorf("failed to initialize audio: %w", err)
	}

	// Convert beep samples to raw PCM data
	audioData, err := p.convertToRawPCM(streamer, format)
	if err != nil {
		return fmt.Errorf("failed to convert audio data: %w", err)
	}

	// Calculate duration
	samplesPerSecond := p.sampleRate * p.channels
	totalSamples := len(audioData) / 2 // 16-bit samples = 2 bytes each
	p.duration = time.Duration(totalSamples/samplesPerSecond) * time.Second

	// Store audio data
	p.audioData = audioData
	p.isLoaded = true
	p.currentFile = filename
	p.position = 0

	return nil
}

// convertToRawPCM converts beep streamer to raw PCM data for Oto
func (p *AudioPlayer) convertToRawPCM(streamer beep.StreamSeekCloser, format beep.Format) ([]byte, error) {
	// Create a buffer to hold all samples
	var samples [][2]float64
	
	// Read all samples from the streamer
	for {
		sampleBuffer := make([][2]float64, 512)
		n, ok := streamer.Stream(sampleBuffer)
		if !ok {
			break
		}
		samples = append(samples, sampleBuffer[:n]...)
	}

	// Convert float64 samples to 16-bit PCM with volume scaling
	pcmData := make([]byte, len(samples)*2*p.channels)
	for i, sample := range samples {
		// Apply volume scaling
		left := sample[0] * p.volume
		right := sample[1] * p.volume

		// Clamp values to prevent distortion
		if left > 1.0 {
			left = 1.0
		} else if left < -1.0 {
			left = -1.0
		}
		if right > 1.0 {
			right = 1.0
		} else if right < -1.0 {
			right = -1.0
		}

		// Convert left channel
		leftInt := int16(left * 32767)
		pcmData[i*4] = byte(leftInt)
		pcmData[i*4+1] = byte(leftInt >> 8)

		// Convert right channel (or duplicate left if mono)
		rightInt := int16(right * 32767)
		if p.channels > 1 {
			pcmData[i*4+2] = byte(rightInt)
			pcmData[i*4+3] = byte(rightInt >> 8)
		}
	}

	return pcmData, nil
}

// Play starts audio playback using Oto with optimized responsiveness
func (p *AudioPlayer) Play() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.isLoaded || len(p.audioData) == 0 {
		return fmt.Errorf("no audio file loaded")
	}

	if p.otoContext == nil {
		return fmt.Errorf("audio context not initialized")
	}

	// Stop any existing playback quickly
	p.stopInternal()

	// Create a new player with the raw PCM data
	p.player = p.otoContext.NewPlayer(bytes.NewReader(p.audioData))
	
	// Start playback immediately
	p.player.Play()
	p.isPlaying = true
	p.isPaused = false
	p.position = 0

	// Create new done channel
	p.playbackDone = make(chan struct{})

	// Start position tracking in background (don't wait)
	go p.trackPosition()

	return nil
}

// trackPosition tracks the playback position
func (p *AudioPlayer) trackPosition() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	startTime := time.Now()

	for {
		select {
		case <-ticker.C:
			p.mutex.Lock()
			if !p.isPlaying || p.isPaused {
				p.mutex.Unlock()
				return
			}

			// Update position based on elapsed time
			elapsed := time.Since(startTime)
			p.position = elapsed

			// Check if playback is finished
			if p.position >= p.duration {
				p.position = p.duration
				p.isPlaying = false
				p.isPaused = false
				close(p.playbackDone)
				p.mutex.Unlock()
				return
			}
			p.mutex.Unlock()

		case <-p.playbackDone:
			return
		}
	}
}

// Pause pauses audio playback
func (p *AudioPlayer) Pause() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.isPlaying && p.player != nil {
		p.player.Pause()
		p.isPaused = true
		p.isPlaying = false
	}
}

// Resume resumes paused audio playback
func (p *AudioPlayer) Resume() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.isPaused && p.player != nil {
		p.player.Play()
		p.isPaused = false
		p.isPlaying = true
		// Note: Position tracking should be handled by the calling application
	}
}

// stopInternal stops playback without mutex (for internal use)
func (p *AudioPlayer) stopInternal() {
	if p.player != nil {
		p.player.Pause()
		p.player.Close()
		p.player = nil
	}
	p.isPlaying = false
	p.isPaused = false
	p.position = 0
}

// Stop stops audio playback
func (p *AudioPlayer) Stop() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.stopInternal()
}

// IsPlaying returns true if audio is currently playing
func (p *AudioPlayer) IsPlaying() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.isPlaying && !p.isPaused
}

// GetPosition returns the current playback position
func (p *AudioPlayer) GetPosition() time.Duration {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.position
}

// GetDuration returns the total duration of the loaded audio
func (p *AudioPlayer) GetDuration() time.Duration {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.duration
}

// WaitForCompletion waits for the current playback to finish
func (p *AudioPlayer) WaitForCompletion() {
	if p.playbackDone != nil {
		<-p.playbackDone
	}
}

// SetVolume sets the audio volume (0.0 to 1.0)
func (p *AudioPlayer) SetVolume(volume float64) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Clamp volume to valid range
	if volume < 0.0 {
		volume = 0.0
	} else if volume > 1.0 {
		volume = 1.0
	}

	p.volume = volume
}

// GetVolume returns the current volume level (0.0 to 1.0)
func (p *AudioPlayer) GetVolume() float64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.volume
}

// SeekTo seeks to a specific position in the current audio file
func (p *AudioPlayer) SeekTo(position time.Duration) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.isLoaded || p.player == nil {
		return fmt.Errorf("no audio file loaded or player not available")
	}

	if position < 0 {
		position = 0
	} else if position > p.duration {
		position = p.duration
	}

	// Calculate the byte position based on the duration
	samplesPerSecond := p.sampleRate * p.channels
	targetSamples := int(position.Seconds()) * samplesPerSecond

	// For Oto v3, seeking requires restarting the player
	// Store the current playback state
	wasPlaying := p.isPlaying

	// Stop current playback
	p.stopInternal()

	// Recreate the player with the original audio data
	p.player = p.otoContext.NewPlayer(bytes.NewReader(p.audioData))

	// Skip to the target position by consuming samples
	sampleSize := 2 * p.channels // 16-bit samples
	bytesToSkip := targetSamples * sampleSize

	// Create a limited reader that skips the initial bytes
	audioReader := bytes.NewReader(p.audioData)
	audioReader.Seek(int64(bytesToSkip), 0)

	// Create a new player starting from the seek position
	p.player = p.otoContext.NewPlayer(audioReader)
	p.position = position

	// Restore playback state
	if wasPlaying {
		p.isPlaying = true
		p.isPaused = false
		p.player.Play()
	}

	return nil
}

// Close cleans up the audio player
func (p *AudioPlayer) Close() error {
	p.Stop()
	// Oto v3 context doesn't need explicit closing
	return nil
}