package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"spotiflac/backend/core"
	"spotiflac/pkg/config"

	"github.com/spf13/cobra"
)

func runAlbumDownload(cmd *cobra.Command, args []string) error {
	spotifyURL := args[0]

	if !isValidSpotifyURL(spotifyURL) {
		return fmt.Errorf("invalid Spotify URL. Expected album URL like: https://open.spotify.com/album/...")
	}

	cfg, err := loadConfig(cmd)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	reporter := NewCliProgressReporter()
	downloader := core.NewAlbumDownloader(cfg, reporter)

	if err := downloader.DownloadAlbum(spotifyURL); err != nil {
		reporter.PrintSummary()
		return fmt.Errorf("download failed: %w", err)
	}

	reporter.PrintSummary()
	return nil
}

func runPlaylistDownload(cmd *cobra.Command, args []string) error {
	spotifyURL := args[0]

	if !isValidPlaylistURL(spotifyURL) {
		return fmt.Errorf("invalid Spotify URL. Expected playlist URL like: https://open.spotify.com/playlist/...")
	}

	cfg, err := loadConfig(cmd)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	reporter := NewCliProgressReporter()
	downloader := core.NewAlbumDownloader(cfg, reporter)

	if err := downloader.DownloadPlaylist(spotifyURL); err != nil {
		reporter.PrintSummary()
		return fmt.Errorf("download failed: %w", err)
	}

	reporter.PrintSummary()
	return nil
}

func runDiscographyDownload(cmd *cobra.Command, args []string) error {
	spotifyURL := args[0]

	if !isValidDiscographyURL(spotifyURL) {
		return fmt.Errorf("invalid Spotify URL. Expected discography URL like: https://open.spotify.com/artist/.../discography/album")
	}

	cfg, err := loadConfig(cmd)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	reporter := NewCliProgressReporter()
	downloader := core.NewAlbumDownloader(cfg, reporter)

	if err := downloader.DownloadDiscography(spotifyURL); err != nil {
		reporter.PrintSummary()
		return fmt.Errorf("download failed: %w", err)
	}

	reporter.PrintSummary()
	return nil
}

func loadConfig(cmd *cobra.Command) (*config.AppConfig, error) {
	configPath, _ := cmd.Flags().GetString("config")
	if configPath == "" {
		configPath = config.GetDefaultConfigPath()
	}

	cfg := config.LoadOrDefault(configPath)

	if output, _ := cmd.Flags().GetString("output"); output != "" {
		if strings.HasPrefix(output, "~") {
			home, _ := os.UserHomeDir()
			output = filepath.Join(home, output[1:])
		}
		cfg.OutputDir = output
	}

	if service, _ := cmd.Flags().GetString("service"); service != "" {
		cfg.PreferredService = service
	}

	if format, _ := cmd.Flags().GetString("format"); format != "" {
		cfg.AudioFormat = strings.ToUpper(format)
	}

	if filenameFormat, _ := cmd.Flags().GetString("filename-format"); filenameFormat != "" {
		cfg.FilenameFormat = filenameFormat
	}

	if noTrackNumbers, _ := cmd.Flags().GetBool("no-track-numbers"); noTrackNumbers {
		cfg.TrackNumbers = false
	}

	if noAlbumFolders, _ := cmd.Flags().GetBool("no-album-folders"); noAlbumFolders {
		cfg.AlbumFolders = false
	}

	if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
		cfg.Verbose = true
	}

	cfg.Validate()
	return cfg, nil
}

func isValidSpotifyURL(url string) bool {
	return strings.Contains(url, "open.spotify.com/album/") ||
		strings.Contains(url, "spotify.com/album/") ||
		strings.Contains(url, "spotify:album:")
}

func isValidPlaylistURL(url string) bool {
	return strings.Contains(url, "open.spotify.com/playlist/") ||
		strings.Contains(url, "spotify.com/playlist/") ||
		strings.Contains(url, "spotify:playlist:")
}

func isValidDiscographyURL(url string) bool {
	return strings.Contains(url, "/artist/") && strings.Contains(url, "/discography/")
}
