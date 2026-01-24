package main

import (
	"fmt"
	"time"
)

type CliProgressReporter struct {
	currentTrack string
	successCount int
	failedCount  int
	skippedCount int
	albumName    string
	trackCount   int
	startTime    time.Time
	lastProgress float64
}

func NewCliProgressReporter() *CliProgressReporter {
	return &CliProgressReporter{
		startTime: time.Now(),
	}
}

func (r *CliProgressReporter) OnAlbumStart(albumName string, trackCount int) {
	r.albumName = albumName
	r.trackCount = trackCount
	fmt.Printf("\n📀 Downloading: %s (%d tracks)\n", albumName, trackCount)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func (r *CliProgressReporter) OnTrackStart(trackName, artistName string) {
	r.currentTrack = trackName
	r.lastProgress = 0
	fmt.Printf("⏳ %s - %s", trackName, artistName)
}

func (r *CliProgressReporter) OnTrackProgress(downloaded, speed float64) {
	if downloaded-r.lastProgress > 0.5 || speed > 0 {
		r.lastProgress = downloaded
		fmt.Printf("\r⏳ %s (%.1f MB @ %.1f MB/s)", r.currentTrack, downloaded, speed)
	}
}

func (r *CliProgressReporter) OnTrackComplete(trackName, filePath string, sizeMB float64) {
	fmt.Printf("\r✓ %s (%.1f MB)                    \n", trackName, sizeMB)
	r.successCount++
}

func (r *CliProgressReporter) OnTrackFailed(trackName, errorMsg string) {
	shortError := errorMsg
	if len(errorMsg) > 50 {
		shortError = errorMsg[:50] + "..."
	}
	fmt.Printf("\r✗ %s - ERROR: %s                    \n", trackName, shortError)
	r.failedCount++
}

func (r *CliProgressReporter) OnTrackSkipped(trackName, reason string) {
	fmt.Printf("\r⚠ %s - SKIPPED: %s                    \n", trackName, reason)
	r.skippedCount++
}

func (r *CliProgressReporter) OnAlbumComplete(successCount, failedCount, skippedCount int) {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func (r *CliProgressReporter) PrintSummary() {
	duration := time.Since(r.startTime)
	fmt.Printf("\n✨ Completed in %s\n", duration.Round(time.Second))

	if r.successCount > 0 {
		fmt.Printf("   ✓ %d track(s) downloaded\n", r.successCount)
	}
	if r.skippedCount > 0 {
		fmt.Printf("   ⚠ %d track(s) skipped\n", r.skippedCount)
	}
	if r.failedCount > 0 {
		fmt.Printf("   ✗ %d track(s) failed\n", r.failedCount)
	}

	fmt.Println()
}
