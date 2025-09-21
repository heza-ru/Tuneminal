package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PerformanceData represents karaoke performance statistics
type PerformanceData struct {
	Date        time.Time `json:"date" csv:"date"`
	SongTitle   string    `json:"song_title" csv:"song_title"`
	Artist      string    `json:"artist" csv:"artist"`
	Score       int       `json:"score" csv:"score"`
	Streak      int       `json:"streak" csv:"streak"`
	Accuracy    float64   `json:"accuracy" csv:"accuracy"`
	Duration    string    `json:"duration" csv:"duration"`
}

// LibraryData represents a song in the music library
type LibraryData struct {
	Title      string `json:"title" csv:"title"`
	Artist     string `json:"artist" csv:"artist"`
	Path       string `json:"path" csv:"path"`
	LyricsPath string `json:"lyrics_path" csv:"lyrics_path"`
	Duration   string `json:"duration" csv:"duration"`
	Format     string `json:"format" csv:"format"`
	Size       int64  `json:"size" csv:"size"`
}

// ExportManager handles data export functionality
type ExportManager struct {
	exportDir string
}

// NewExportManager creates a new export manager
func NewExportManager() *ExportManager {
	homeDir, _ := os.UserHomeDir()
	exportDir := filepath.Join(homeDir, ".tuneminal", "exports")

	return &ExportManager{
		exportDir: exportDir,
	}
}

// ExportPerformanceData exports karaoke performance statistics
func (em *ExportManager) ExportPerformanceData(performances []PerformanceData, format string) error {
	// Create export directory if it doesn't exist
	if err := os.MkdirAll(em.exportDir, 0755); err != nil {
		return err
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("karaoke_performance_%s.%s", timestamp, format)
	filepath := filepath.Join(em.exportDir, filename)

	switch format {
	case "json":
		return em.exportPerformanceAsJSON(performances, filepath)
	case "csv":
		return em.exportPerformanceAsCSV(performances, filepath)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// ExportLibraryData exports music library information
func (em *ExportManager) ExportLibraryData(library []LibraryData, format string) error {
	// Create export directory if it doesn't exist
	if err := os.MkdirAll(em.exportDir, 0755); err != nil {
		return err
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("music_library_%s.%s", timestamp, format)
	filepath := filepath.Join(em.exportDir, filename)

	switch format {
	case "json":
		return em.exportLibraryAsJSON(library, filepath)
	case "csv":
		return em.exportLibraryAsCSV(library, filepath)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// exportPerformanceAsJSON exports performance data as JSON
func (em *ExportManager) exportPerformanceAsJSON(performances []PerformanceData, filepath string) error {
	data := map[string]interface{}{
		"export_date": time.Now(),
		"total_sessions": len(performances),
		"performances": performances,
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// exportPerformanceAsCSV exports performance data as CSV
func (em *ExportManager) exportPerformanceAsCSV(performances []PerformanceData, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"date", "song_title", "artist", "score", "streak", "accuracy", "duration"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for _, perf := range performances {
		record := []string{
			perf.Date.Format("2006-01-02 15:04:05"),
			perf.SongTitle,
			perf.Artist,
			fmt.Sprintf("%d", perf.Score),
			fmt.Sprintf("%d", perf.Streak),
			fmt.Sprintf("%.1f%%", perf.Accuracy),
			perf.Duration,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// exportLibraryAsJSON exports library data as JSON
func (em *ExportManager) exportLibraryAsJSON(library []LibraryData, filepath string) error {
	data := map[string]interface{}{
		"export_date": time.Now(),
		"total_songs": len(library),
		"library": library,
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// exportLibraryAsCSV exports library data as CSV
func (em *ExportManager) exportLibraryAsCSV(library []LibraryData, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"title", "artist", "path", "lyrics_path", "duration", "format", "size"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for _, song := range library {
		record := []string{
			song.Title,
			song.Artist,
			song.Path,
			song.LyricsPath,
			song.Duration,
			song.Format,
			fmt.Sprintf("%d", song.Size),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// ListExports returns a list of all exported files
func (em *ExportManager) ListExports() ([]string, error) {
	// Create export directory if it doesn't exist
	if err := os.MkdirAll(em.exportDir, 0755); err != nil {
		return nil, err
	}

	files, err := os.ReadDir(em.exportDir)
	if err != nil {
		return nil, err
	}

	var exports []string
	for _, file := range files {
		exports = append(exports, file.Name())
	}

	return exports, nil
}

// GetExportPath returns the full path to the export directory
func (em *ExportManager) GetExportPath() string {
	return em.exportDir
}
