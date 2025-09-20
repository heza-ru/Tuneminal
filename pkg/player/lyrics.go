package player

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// LoadLyrics loads lyrics from an LRC file
func LoadLyrics(filename string) ([]LyricEntry, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open lyrics file: %w", err)
	}
	defer file.Close()

	var lyrics []LyricEntry
	scanner := bufio.NewScanner(file)
	
	// Regex to match LRC time tags [mm:ss.xx] or [mm:ss]
	timeRegex := regexp.MustCompile(`\[(\d{2}):(\d{2})(?:\.(\d{2}))?\]`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Find all time tags in the line
		matches := timeRegex.FindAllStringSubmatch(line, -1)
		if len(matches) == 0 {
			continue
		}

		// Extract the text part (after all time tags)
		text := timeRegex.ReplaceAllString(line, "")
		text = strings.TrimSpace(text)
		
		if text == "" {
			continue
		}

		// Parse each time tag and create a lyric entry
		for _, match := range matches {
			if len(match) < 3 {
				continue
			}

			minutes, err := strconv.Atoi(match[1])
			if err != nil {
				continue
			}

			seconds, err := strconv.Atoi(match[2])
			if err != nil {
				continue
			}

			// Parse centiseconds if present
			centiseconds := 0
			if len(match) > 3 && match[3] != "" {
				centiseconds, err = strconv.Atoi(match[3])
				if err != nil {
					centiseconds = 0
				}
			}

			// Calculate total time
			totalSeconds := time.Duration(minutes)*time.Minute + 
				time.Duration(seconds)*time.Second + 
				time.Duration(centiseconds)*10*time.Millisecond

			lyrics = append(lyrics, LyricEntry{
				Time: totalSeconds,
				Text: text,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading lyrics file: %w", err)
	}

	return lyrics, nil
}

// ParseLRCHeader parses LRC header information (optional)
func ParseLRCHeader(line string) map[string]string {
	headerRegex := regexp.MustCompile(`\[([^:]+):([^\]]+)\]`)
	matches := headerRegex.FindAllStringSubmatch(line, -1)
	
	headers := make(map[string]string)
	for _, match := range matches {
		if len(match) >= 3 {
			key := strings.ToLower(strings.TrimSpace(match[1]))
			value := strings.TrimSpace(match[2])
			headers[key] = value
		}
	}
	
	return headers
}

// ValidateLyrics checks if lyrics are properly formatted
func ValidateLyrics(lyrics []LyricEntry) error {
	if len(lyrics) == 0 {
		return fmt.Errorf("no lyrics found")
	}

	// Check if lyrics are in chronological order
	for i := 1; i < len(lyrics); i++ {
		if lyrics[i].Time < lyrics[i-1].Time {
			return fmt.Errorf("lyrics are not in chronological order at position %d", i)
		}
	}

	return nil
}

// FindLyricAtTime finds the lyric entry at or before the given time
func FindLyricAtTime(lyrics []LyricEntry, targetTime time.Duration) *LyricEntry {
	var current *LyricEntry
	
	for i := range lyrics {
		if lyrics[i].Time <= targetTime {
			current = &lyrics[i]
		} else {
			break
		}
	}
	
	return current
}


