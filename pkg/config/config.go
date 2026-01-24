package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	OutputDir        string `yaml:"output_dir"`
	AudioFormat      string `yaml:"audio_format"`
	PreferredService string `yaml:"preferred_service"`
	FilenameFormat   string `yaml:"filename_format"`
	TrackNumbers     bool   `yaml:"track_numbers"`
	AlbumFolders     bool   `yaml:"album_folders"`
	Verbose          bool   `yaml:"verbose"`
}

func (c *AppConfig) GetOutputDir() string {
	return c.OutputDir
}

func (c *AppConfig) GetAudioFormat() string {
	return c.AudioFormat
}

func (c *AppConfig) GetPreferredService() string {
	return c.PreferredService
}

func (c *AppConfig) GetFilenameFormat() string {
	return c.FilenameFormat
}

func (c *AppConfig) UseTrackNumbers() bool {
	return c.TrackNumbers
}

func (c *AppConfig) CreateAlbumFolders() bool {
	return c.AlbumFolders
}

func (c *AppConfig) IsVerbose() bool {
	return c.Verbose
}

func (c *AppConfig) Validate() {
	if c.OutputDir == "" {
		home, _ := os.UserHomeDir()
		c.OutputDir = filepath.Join(home, "Music", "SpotiFLAC")
	}
	if strings.HasPrefix(c.OutputDir, "~") {
		home, _ := os.UserHomeDir()
		c.OutputDir = filepath.Join(home, c.OutputDir[1:])
	}

	if c.AudioFormat == "" {
		c.AudioFormat = "LOSSLESS"
	}

	if c.PreferredService == "" {
		c.PreferredService = "tidal"
	}

	if c.FilenameFormat == "" {
		c.FilenameFormat = "title-artist"
	}
}

func DefaultConfig() *AppConfig {
	home, _ := os.UserHomeDir()
	return &AppConfig{
		OutputDir:        filepath.Join(home, "Music", "SpotiFLAC"),
		AudioFormat:      "LOSSLESS",
		PreferredService: "tidal",
		FilenameFormat:   "title-artist",
		TrackNumbers:     true,
		AlbumFolders:     true,
	}
}

func GetDefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".spotiflac", "config.yaml")
}

func LoadOrDefault(path string) *AppConfig {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return DefaultConfig()
	}

	return cfg
}
