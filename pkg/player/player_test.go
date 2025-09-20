package player

import (
	"testing"
	"time"
)

func TestNewAudioPlayer(t *testing.T) {
	player := NewAudioPlayer()
	
	if player == nil {
		t.Fatal("NewAudioPlayer() returned nil")
	}
	
	if player.done == nil {
		t.Error("Audio player done channel is nil")
	}
}

func TestGetAudioSamples(t *testing.T) {
	player := NewAudioPlayer()
	
	samples := player.GetAudioSamples()
	
	if len(samples) == 0 {
		t.Error("GetAudioSamples() returned empty slice")
	}
	
	if len(samples) != 1024 {
		t.Errorf("Expected 1024 samples, got %d", len(samples))
	}
	
	// Check that samples are within expected range
	for i, sample := range samples {
		if sample < -1.0 || sample > 1.0 {
			t.Errorf("Sample %d is out of range: %f", i, sample)
		}
	}
}

func TestLyricEntry(t *testing.T) {
	lyric := LyricEntry{
		Time: 30 * time.Second,
		Text: "Test lyric",
	}
	
	if lyric.Time != 30*time.Second {
		t.Error("Lyric time not set correctly")
	}
	
	if lyric.Text != "Test lyric" {
		t.Error("Lyric text not set correctly")
	}
}


