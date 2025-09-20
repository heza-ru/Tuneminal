package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// ScanDemoFiles scans the uploads/demo directory for audio and lyrics files
func ScanDemoFiles() ([]string, []string) {
	demoDir := "uploads/demo"
	
	var songFiles []string
	var lyricsFiles []string

	// Check if demo directory exists
	if _, err := os.Stat(demoDir); os.IsNotExist(err) {
		return songFiles, lyricsFiles
	}

	// Walk through the demo directory
	err := filepath.Walk(demoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		
		// Check for audio files
		if isAudioFile(ext) {
			songFiles = append(songFiles, path)
		}
		
		// Check for lyrics files
		if isLyricsFile(ext) {
			lyricsFiles = append(lyricsFiles, path)
		}

		return nil
	})

	if err != nil {
		// If there's an error, return empty slices
		return []string{}, []string{}
	}

	return songFiles, lyricsFiles
}

// isAudioFile checks if the file extension is a supported audio format
func isAudioFile(ext string) bool {
	supportedAudio := map[string]bool{
		".mp3": true,
		".wav": true,
		".m4a": true,
		".flac": true,
	}
	
	return supportedAudio[ext]
}

// isLyricsFile checks if the file extension is a supported lyrics format
func isLyricsFile(ext string) bool {
	supportedLyrics := map[string]bool{
		".lrc": true,
		".txt": true,
	}
	
	return supportedLyrics[ext]
}

// GetFileInfo returns basic information about a file
func GetFileInfo(filename string) (os.FileInfo, error) {
	return os.Stat(filename)
}

// EnsureDir ensures that a directory exists, creating it if necessary
func EnsureDir(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0755)
	}
	return nil
}

// GetFileNameWithoutExt returns the filename without its extension
func GetFileNameWithoutExt(filename string) string {
	return strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
}


