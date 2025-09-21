package playlist

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Playlist represents a music playlist
type Playlist struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
	Songs       []string  `json:"songs"` // Song paths
}

// PlaylistManager manages playlist operations
type PlaylistManager struct {
	playlistDir string
}

// NewPlaylistManager creates a new playlist manager
func NewPlaylistManager() *PlaylistManager {
	homeDir, _ := os.UserHomeDir()
	playlistDir := filepath.Join(homeDir, ".tuneminal", "playlists")

	return &PlaylistManager{
		playlistDir: playlistDir,
	}
}

// CreatePlaylist creates a new playlist
func (pm *PlaylistManager) CreatePlaylist(name, description string) (*Playlist, error) {
	now := time.Now()

	playlist := &Playlist{
		Name:        name,
		Description: description,
		Created:     now,
		Modified:    now,
		Songs:       []string{},
	}

	// Save the playlist
	return playlist, pm.SavePlaylist(playlist)
}

// LoadPlaylist loads a playlist by name
func (pm *PlaylistManager) LoadPlaylist(name string) (*Playlist, error) {
	filename := fmt.Sprintf("%s.json", name)
	filepath := filepath.Join(pm.playlistDir, filename)

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var playlist Playlist
	if err := json.Unmarshal(data, &playlist); err != nil {
		return nil, err
	}

	return &playlist, nil
}

// SavePlaylist saves a playlist to file
func (pm *PlaylistManager) SavePlaylist(playlist *Playlist) error {
	// Create playlist directory if it doesn't exist
	if err := os.MkdirAll(pm.playlistDir, 0755); err != nil {
		return err
	}

	// Update modified time
	playlist.Modified = time.Now()

	filename := fmt.Sprintf("%s.json", playlist.Name)
	filepath := filepath.Join(pm.playlistDir, filename)

	data, err := json.MarshalIndent(playlist, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

// DeletePlaylist deletes a playlist
func (pm *PlaylistManager) DeletePlaylist(name string) error {
	filename := fmt.Sprintf("%s.json", name)
	filepath := filepath.Join(pm.playlistDir, filename)
	return os.Remove(filepath)
}

// ListPlaylists returns all available playlists
func (pm *PlaylistManager) ListPlaylists() ([]string, error) {
	// Create playlist directory if it doesn't exist
	if err := os.MkdirAll(pm.playlistDir, 0755); err != nil {
		return nil, err
	}

	files, err := os.ReadDir(pm.playlistDir)
	if err != nil {
		return nil, err
	}

	var playlists []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			name := file.Name()[:len(file.Name())-5] // Remove .json extension
			playlists = append(playlists, name)
		}
	}

	return playlists, nil
}

// AddSongToPlaylist adds a song to a playlist
func (pm *PlaylistManager) AddSongToPlaylist(playlistName, songPath string) error {
	playlist, err := pm.LoadPlaylist(playlistName)
	if err != nil {
		return err
	}

	// Check if song is already in playlist
	for _, song := range playlist.Songs {
		if song == songPath {
			return fmt.Errorf("song already exists in playlist")
		}
	}

	playlist.Songs = append(playlist.Songs, songPath)
	return pm.SavePlaylist(playlist)
}

// RemoveSongFromPlaylist removes a song from a playlist
func (pm *PlaylistManager) RemoveSongFromPlaylist(playlistName, songPath string) error {
	playlist, err := pm.LoadPlaylist(playlistName)
	if err != nil {
		return err
	}

	for i, song := range playlist.Songs {
		if song == songPath {
			playlist.Songs = append(playlist.Songs[:i], playlist.Songs[i+1:]...)
			return pm.SavePlaylist(playlist)
		}
	}

	return fmt.Errorf("song not found in playlist")
}

// GetPlaylistSongs returns all songs in a playlist
func (pm *PlaylistManager) GetPlaylistSongs(playlistName string) ([]string, error) {
	playlist, err := pm.LoadPlaylist(playlistName)
	if err != nil {
		return nil, err
	}

	return playlist.Songs, nil
}
