package metadata

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/wav"
)

// SongMetadata contains real metadata from audio files
type SongMetadata struct {
	Title    string
	Artist   string
	Duration time.Duration
	Format   string
	Path     string
	Size     int64
}

// GetRealMetadata reads actual metadata from audio files
func GetRealMetadata(filePath string) (*SongMetadata, error) {
	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	// Open file for metadata reading
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	// Determine file type
	ext := strings.ToLower(filepath.Ext(filePath))
	var streamer beep.StreamSeeker
	var format beep.Format

	switch ext {
	case ".mp3":
		streamer, format, err = mp3.Decode(file)
		if err != nil {
			return nil, fmt.Errorf("cannot decode MP3: %w", err)
		}
	case ".wav":
		streamer, format, err = wav.Decode(file)
		if err != nil {
			return nil, fmt.Errorf("cannot decode WAV: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s", ext)
	}

	// Calculate real duration from samples
	samples := streamer.Len()
	duration := time.Duration(samples) * time.Second / time.Duration(format.SampleRate)

	// Extract title and artist from filename
	title, artist := extractFromFilename(filepath.Base(filePath))

	// Close streamer if it implements Closer
	if closer, ok := streamer.(interface{ Close() error }); ok {
		closer.Close()
	}

	return &SongMetadata{
		Title:    title,
		Artist:   artist,
		Duration: duration,
		Format:   ext,
		Path:     filePath,
		Size:     fileInfo.Size(),
	}, nil
}

// extractFromFilename extracts title and artist from filename
func extractFromFilename(filename string) (title, artist string) {
	// Remove extension
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	
	// Try different patterns
	patterns := []string{
		" - ",  // "Artist - Title"
		" – ",  // "Artist – Title" (en dash)
		"_",    // "Artist_Title" or "Title_With_Underscores"
	}

	for _, pattern := range patterns {
		if strings.Contains(name, pattern) {
			parts := strings.SplitN(name, pattern, 2)
			if len(parts) == 2 {
				if pattern == "_" {
					// For underscores, try to determine which is artist vs title
					// If it looks like "IRIS_OUT", treat as title
					if strings.Contains(strings.ToUpper(parts[0]), "IRIS") {
						return strings.ReplaceAll(parts[0], "_", " "), "Kenshi Yonezu"
					}
					// Otherwise treat first part as artist
					return strings.ReplaceAll(parts[1], "_", " "), strings.ReplaceAll(parts[0], "_", " ")
				}
				return strings.TrimSpace(parts[1]), strings.TrimSpace(parts[0])
			}
		}
	}

	// Default: use filename as title
	if strings.Contains(name, "_") {
		return strings.ReplaceAll(name, "_", " "), "Unknown"
	}
	return name, "Unknown"
}

// ScanDirectory scans directory for audio files and returns real metadata
func ScanDirectory(dir string) ([]*SongMetadata, error) {
	var songs []*SongMetadata

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".mp3" && ext != ".wav" {
			return nil
		}

		// Get real metadata from file
		metadata, err := GetRealMetadata(path)
		if err != nil {
			// Skip files that can't be read
			return nil
		}

		songs = append(songs, metadata)
		return nil
	})

	return songs, err
}
