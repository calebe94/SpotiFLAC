package core

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"spotiflac/backend"
)

func isValidISRC(isrc string) bool {
	if len(isrc) != 12 {
		return false
	}
	matched, _ := regexp.MatchString(`^[A-Z]{2}[A-Z0-9]{3}\d{7}$`, isrc)
	return matched
}

type AlbumDownloader struct {
	config   Config
	reporter ProgressReporter
	fetcher  *MetadataFetcher
}

func NewAlbumDownloader(config Config, reporter ProgressReporter) *AlbumDownloader {
	if reporter == nil {
		reporter = &NoOpProgressReporter{}
	}
	return &AlbumDownloader{
		config:   config,
		reporter: reporter,
		fetcher:  NewMetadataFetcher(),
	}
}

func (d *AlbumDownloader) DownloadAlbum(spotifyURL string) error {
	album, err := d.fetcher.FetchAlbum(spotifyURL)
	if err != nil {
		return fmt.Errorf("failed to fetch album metadata: %w", err)
	}

	outputDir := d.config.GetOutputDir()
	if d.config.CreateAlbumFolders() {
		albumFolder := backend.SanitizeFolderPath(fmt.Sprintf("%s - %s", album.Artist, album.Name))
		outputDir = filepath.Join(outputDir, albumFolder)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	d.reporter.OnAlbumStart(album.Name, album.TrackCount)

	successCount := 0
	failedCount := 0
	skippedCount := 0

	for _, track := range album.Tracks {
		result := d.downloadTrack(track, outputDir)

		switch result.Status {
		case DownloadSuccess:
			successCount++
		case DownloadFailed:
			failedCount++
		case DownloadSkipped:
			skippedCount++
		}
	}

	d.reporter.OnAlbumComplete(successCount, failedCount, skippedCount)

	if failedCount > 0 && successCount == 0 {
		return fmt.Errorf("all tracks failed to download")
	}

	return nil
}

type DownloadStatus int

const (
	DownloadSuccess DownloadStatus = iota
	DownloadFailed
	DownloadSkipped
)

type DownloadResult struct {
	Status   DownloadStatus
	FilePath string
	Error    error
	SizeMB   float64
}

func (d *AlbumDownloader) downloadTrack(track TrackMetadata, outputDir string) DownloadResult {
	d.reporter.OnTrackStart(track.Name, track.Artist)

	preferredService := d.config.GetPreferredService()
	services := []string{preferredService}

	allServices := []string{"tidal", "amazon", "qobuz"}
	for _, svc := range allServices {
		if svc != preferredService {
			services = append(services, svc)
		}
	}

	if d.config.IsVerbose() {
		fmt.Printf("  [VERBOSE] Trying services in order: %v\n", services)
	}

	var lastErr error
	serviceErrors := make(map[string]string)

	for _, service := range services {
		if d.config.IsVerbose() {
			fmt.Printf("  [VERBOSE] Attempting %s for track: %s\n", service, track.Name)
		}
		result := d.downloadTrackFromService(track, outputDir, service)

		if result.Status == DownloadSuccess {
			if d.config.IsVerbose() {
				fmt.Printf("  [VERBOSE] ✓ Successfully downloaded from %s\n", service)
			}
			d.reporter.OnTrackComplete(track.Name, result.FilePath, result.SizeMB)
			return result
		}

		if result.Status == DownloadSkipped {
			d.reporter.OnTrackSkipped(track.Name, "file already exists")
			return result
		}

		if d.config.IsVerbose() {
			fmt.Printf("  [VERBOSE] ✗ %s failed: %v\n", service, result.Error)
		}
		serviceErrors[service] = result.Error.Error()
		lastErr = result.Error
	}

	var errorMsg string
	if d.config.IsVerbose() {
		errorMsg = fmt.Sprintf("all services failed. Tried: %v", services)
		for svc, err := range serviceErrors {
			errorMsg += fmt.Sprintf("\n  - %s: %s", svc, err)
		}
	} else {
		if lastErr != nil {
			errorMsg = fmt.Sprintf("not available on %s (tried: %s, %s, %s)",
				preferredService, services[0], services[1], services[2])
		} else {
			errorMsg = fmt.Sprintf("not available on any service (tried: %s, %s, %s)",
				services[0], services[1], services[2])
		}
	}

	if d.config.IsVerbose() {
		fmt.Printf("  [VERBOSE] ✗ All services failed. Last error: %s\n", errorMsg)
	}
	d.reporter.OnTrackFailed(track.Name, errorMsg)

	return DownloadResult{
		Status: DownloadFailed,
		Error:  lastErr,
	}
}

func (d *AlbumDownloader) downloadTrackFromService(track TrackMetadata, outputDir, service string) DownloadResult {
	var filename string
	var err error

	audioFormat := d.config.GetAudioFormat()
	filenameFormat := d.config.GetFilenameFormat()
	useTrackNumbers := d.config.UseTrackNumbers()

	switch service {
	case "amazon":
		downloader := backend.NewAmazonDownloader()
		if track.SpotifyID == "" {
			return DownloadResult{Status: DownloadFailed, Error: fmt.Errorf("spotify ID required for Amazon")}
		}
		filename, err = downloader.DownloadBySpotifyID(
			track.SpotifyID,
			outputDir,
			audioFormat,
			filenameFormat,
			useTrackNumbers,
			track.TrackNumber,
			track.Name,
			track.Artist,
			track.AlbumName,
			track.AlbumArtist,
			track.ReleaseDate,
			track.Images,
			track.TrackNumber,
			track.DiscNumber,
			track.TotalTracks,
			false,
			1,
			"",
			"",
			fmt.Sprintf("https://open.spotify.com/track/%s", track.SpotifyID),
		)

	case "tidal":
		downloader := backend.NewTidalDownloader("")
		if track.SpotifyID == "" {
			return DownloadResult{Status: DownloadFailed, Error: fmt.Errorf("spotify ID required for Tidal")}
		}
		filename, err = downloader.Download(
			track.SpotifyID,
			outputDir,
			audioFormat,
			filenameFormat,
			useTrackNumbers,
			track.TrackNumber,
			track.Name,
			track.Artist,
			track.AlbumName,
			track.AlbumArtist,
			track.ReleaseDate,
			true,
			track.Images,
			false,
			track.TrackNumber,
			track.DiscNumber,
			track.TotalTracks,
			1,
			"",
			"",
			fmt.Sprintf("https://open.spotify.com/track/%s", track.SpotifyID),
		)

	case "qobuz":
		isrc := track.ISRC
		if !isValidISRC(isrc) && track.SpotifyID != "" {
			songlinkClient := backend.NewSongLinkClient()
			deezerURL, deezerErr := songlinkClient.GetDeezerURLFromSpotify(track.SpotifyID)
			if deezerErr == nil {
				isrc, _ = backend.GetDeezerISRC(deezerURL)
			}
		}
		if !isValidISRC(isrc) {
			return DownloadResult{Status: DownloadFailed, Error: fmt.Errorf("could not get valid ISRC for Qobuz")}
		}
		downloader := backend.NewQobuzDownloader()
		filename, err = downloader.DownloadByISRC(
			isrc,
			outputDir,
			audioFormat,
			filenameFormat,
			useTrackNumbers,
			track.TrackNumber,
			track.Name,
			track.Artist,
			track.AlbumName,
			track.AlbumArtist,
			track.ReleaseDate,
			true,
			track.Images,
			false,
			track.TrackNumber,
			track.DiscNumber,
			track.TotalTracks,
			1,
			"",
			"",
			fmt.Sprintf("https://open.spotify.com/track/%s", track.SpotifyID),
		)

	default:
		return DownloadResult{
			Status: DownloadFailed,
			Error:  fmt.Errorf("unsupported service: %s", service),
		}
	}

	if err != nil {
		if filename != "" && !strings.HasPrefix(filename, "EXISTS:") {
			if _, statErr := os.Stat(filename); statErr == nil {
				os.Remove(filename)
			}
		}
		return DownloadResult{
			Status: DownloadFailed,
			Error:  err,
		}
	}

	alreadyExists := strings.HasPrefix(filename, "EXISTS:")
	if alreadyExists {
		filename = strings.TrimPrefix(filename, "EXISTS:")

		var sizeMB float64
		if fileInfo, statErr := os.Stat(filename); statErr == nil {
			sizeMB = float64(fileInfo.Size()) / (1024 * 1024)
		}

		return DownloadResult{
			Status:   DownloadSkipped,
			FilePath: filename,
			SizeMB:   sizeMB,
		}
	}

	var sizeMB float64
	if fileInfo, statErr := os.Stat(filename); statErr == nil {
		sizeMB = float64(fileInfo.Size()) / (1024 * 1024)
	}

	return DownloadResult{
		Status:   DownloadSuccess,
		FilePath: filename,
		SizeMB:   sizeMB,
	}
}

func (d *AlbumDownloader) DownloadPlaylist(spotifyURL string) error {
	playlist, err := d.fetcher.FetchPlaylist(spotifyURL)
	if err != nil {
		return fmt.Errorf("failed to fetch playlist metadata: %w", err)
	}

	outputDir := d.config.GetOutputDir()
	if d.config.CreateAlbumFolders() {
		playlistFolder := backend.SanitizeFolderPath(fmt.Sprintf("Playlist - %s", playlist.Name))
		outputDir = filepath.Join(outputDir, playlistFolder)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	d.reporter.OnAlbumStart(playlist.Name, playlist.TrackCount)

	successCount := 0
	failedCount := 0
	skippedCount := 0

	for _, track := range playlist.Tracks {
		result := d.downloadTrack(track, outputDir)

		switch result.Status {
		case DownloadSuccess:
			successCount++
		case DownloadFailed:
			failedCount++
		case DownloadSkipped:
			skippedCount++
		}
	}

	d.reporter.OnAlbumComplete(successCount, failedCount, skippedCount)

	if failedCount > 0 && successCount == 0 {
		return fmt.Errorf("all tracks failed to download")
	}

	return nil
}

func (d *AlbumDownloader) DownloadDiscography(spotifyURL string) error {
	discography, err := d.fetcher.FetchDiscography(spotifyURL)
	if err != nil {
		return fmt.Errorf("failed to fetch discography metadata: %w", err)
	}

	baseOutputDir := d.config.GetOutputDir()
	if d.config.CreateAlbumFolders() {
		artistFolder := backend.SanitizeFolderPath(fmt.Sprintf("%s - Discography", discography.ArtistName))
		baseOutputDir = filepath.Join(baseOutputDir, artistFolder)
	}

	if err := os.MkdirAll(baseOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	totalTracks := len(discography.AllTracks)
	d.reporter.OnAlbumStart(fmt.Sprintf("%s - %s", discography.ArtistName, discography.DiscographyType), totalTracks)

	totalSuccessCount := 0
	totalFailedCount := 0
	totalSkippedCount := 0

	for albumIdx, album := range discography.Albums {
		albumFolder := backend.SanitizeFolderPath(fmt.Sprintf("%s - %s", album.Artist, album.Name))
		albumOutputDir := filepath.Join(baseOutputDir, albumFolder)

		if err := os.MkdirAll(albumOutputDir, 0755); err != nil {
			d.reporter.OnTrackFailed(album.Name, fmt.Sprintf("failed to create album folder: %v", err))
			totalFailedCount += len(album.Tracks)
			continue
		}

		for _, track := range album.Tracks {
			result := d.downloadTrack(track, albumOutputDir)

			switch result.Status {
			case DownloadSuccess:
				totalSuccessCount++
			case DownloadFailed:
				totalFailedCount++
			case DownloadSkipped:
				totalSkippedCount++
			}
		}

		_ = albumIdx
	}

	d.reporter.OnAlbumComplete(totalSuccessCount, totalFailedCount, totalSkippedCount)

	if totalFailedCount > 0 && totalSuccessCount == 0 {
		return fmt.Errorf("all tracks failed to download")
	}

	return nil
}
