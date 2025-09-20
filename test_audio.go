package main

import (
	"fmt"
	"os"

	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

func main() {
	// Initialize speaker
	speaker.Init(44100, 44100/10)
	
	// Test files
	files := []string{
		"uploads/demo/demo_song.mp3",
		"uploads/demo/Kenshi Yonezu - IRIS OUT.mp3",
	}
	
	for _, filename := range files {
		fmt.Printf("Testing file: %s\n", filename)
		
		// Check if file exists
		if _, err := os.Stat(filename); err != nil {
			fmt.Printf("  File does not exist: %v\n", err)
			continue
		}
		
		// Get file size
		info, _ := os.Stat(filename)
		fmt.Printf("  File size: %d bytes\n", info.Size())
		
		// Try to open and decode
		file, err := os.Open(filename)
		if err != nil {
			fmt.Printf("  Cannot open file: %v\n", err)
			continue
		}
		
		streamer, format, err := mp3.Decode(file)
		if err != nil {
			fmt.Printf("  Cannot decode MP3: %v\n", err)
			file.Close()
			continue
		}
		
		fmt.Printf("  Successfully decoded! Format: %v\n", format)
		fmt.Printf("  Duration: %v\n", streamer.Len())
		
		file.Close()
		streamer.Close()
	}
}
