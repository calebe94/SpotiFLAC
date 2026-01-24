package core

type ProgressReporter interface {
	OnAlbumStart(albumName string, trackCount int)
	OnTrackStart(trackName, artistName string)
	OnTrackProgress(downloaded, speed float64)
	OnTrackComplete(trackName, filePath string, sizeMB float64)
	OnTrackFailed(trackName, errorMsg string)
	OnTrackSkipped(trackName, reason string)
	OnAlbumComplete(successCount, failedCount, skippedCount int)
}

type Config interface {
	GetOutputDir() string
	GetAudioFormat() string
	GetPreferredService() string
	GetFilenameFormat() string
	UseTrackNumbers() bool
	CreateAlbumFolders() bool
	IsVerbose() bool
}

type NoOpProgressReporter struct{}

func (n *NoOpProgressReporter) OnAlbumStart(albumName string, trackCount int)               {}
func (n *NoOpProgressReporter) OnTrackStart(trackName, artistName string)                   {}
func (n *NoOpProgressReporter) OnTrackProgress(downloaded, speed float64)                   {}
func (n *NoOpProgressReporter) OnTrackComplete(trackName, filePath string, sizeMB float64)  {}
func (n *NoOpProgressReporter) OnTrackFailed(trackName, errorMsg string)                    {}
func (n *NoOpProgressReporter) OnTrackSkipped(trackName, reason string)                     {}
func (n *NoOpProgressReporter) OnAlbumComplete(successCount, failedCount, skippedCount int) {}
