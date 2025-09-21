package lyrics

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// LyricEditor handles lyrics editing functionality
type LyricEditor struct {
	lines []LyricLine
}

// LyricLine represents a single lyric line with timing
type LyricLine struct {
	Time    time.Duration
	Text    string
	Index   int
}

// NewLyricEditor creates a new lyrics editor
func NewLyricEditor() *LyricEditor {
	return &LyricEditor{
		lines: make([]LyricLine, 0),
	}
}

// LoadLyricsFromFile loads lyrics from an LRC file
func (le *LyricEditor) LoadLyricsFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	le.lines = []LyricLine{}
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

			le.lines = append(le.lines, LyricLine{
				Time:  time,
				Text:  text,
				Index: index,
			})
			index++
		}
	}

	return nil
}

// AddLyricLine adds a new lyric line at the specified time
func (le *LyricEditor) AddLyricLine(time time.Duration, text string) {
	le.lines = append(le.lines, LyricLine{
		Time:  time,
		Text:  text,
		Index: len(le.lines),
	})
}

// UpdateLyricLine updates an existing lyric line
func (le *LyricEditor) UpdateLyricLine(index int, time time.Duration, text string) error {
	if index < 0 || index >= len(le.lines) {
		return fmt.Errorf("line index out of range")
	}

	le.lines[index].Time = time
	le.lines[index].Text = text
	return nil
}

// DeleteLyricLine removes a lyric line
func (le *LyricEditor) DeleteLyricLine(index int) error {
	if index < 0 || index >= len(le.lines) {
		return fmt.Errorf("line index out of range")
	}

	le.lines = append(le.lines[:index], le.lines[index+1:]...)

	// Update indices
	for i := index; i < len(le.lines); i++ {
		le.lines[i].Index = i
	}

	return nil
}

// GetLyricsLines returns all lyric lines
func (le *LyricEditor) GetLyricsLines() []LyricLine {
	return le.lines
}

// SaveLyricsToFile saves lyrics to an LRC file
func (le *LyricEditor) SaveLyricsToFile(filename string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Write header
	writer.WriteString("[ti:Custom Lyrics]\n")
	writer.WriteString("[ar:Unknown Artist]\n")
	writer.WriteString("[al:Unknown Album]\n")
	writer.WriteString("\n")

	// Write lyrics
	for _, line := range le.lines {
		if line.Text != "" {
			minutes := int(line.Time.Minutes())
			seconds := int(line.Time.Seconds()) % 60
			centiseconds := int(line.Time.Milliseconds()) % 1000 / 10

			timeStr := fmt.Sprintf("[%02d:%02d.%02d]", minutes, seconds, centiseconds)
			writer.WriteString(timeStr + line.Text + "\n")
		}
	}

	return writer.Flush()
}

// FormatDuration formats duration for display
func FormatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// ParseTime parses time string in mm:ss.xx format
func ParseTime(timeStr string) (time.Duration, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid time format")
	}

	minutes, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}

	subparts := strings.Split(parts[1], ".")
	if len(subparts) != 2 {
		return 0, fmt.Errorf("invalid time format")
	}

	seconds, err := strconv.Atoi(subparts[0])
	if err != nil {
		return 0, err
	}

	centiseconds, err := strconv.Atoi(subparts[1])
	if err != nil {
		return 0, err
	}

	return time.Duration(minutes)*time.Minute +
		   time.Duration(seconds)*time.Second +
		   time.Duration(centiseconds)*10*time.Millisecond, nil
}
