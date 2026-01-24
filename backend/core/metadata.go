package core

import (
	"context"
	"fmt"
	"time"

	"spotiflac/backend"
)

type TrackMetadata struct {
	ISRC        string
	SpotifyID   string
	Name        string
	Artist      string
	AlbumArtist string
	AlbumName   string
	TrackNumber int
	DiscNumber  int
	TotalTracks int
	Duration    int
	Images      string
	ReleaseDate string
}

type AlbumMetadata struct {
	Name        string
	Artist      string
	ReleaseDate string
	Images      string
	TrackCount  int
	Tracks      []TrackMetadata
}

type PlaylistMetadata struct {
	Name       string
	Owner      string
	TrackCount int
	Tracks     []TrackMetadata
}

type DiscographyMetadata struct {
	ArtistName      string
	DiscographyType string
	TotalAlbums     int
	Albums          []AlbumMetadata
	AllTracks       []TrackMetadata
}

type MetadataFetcher struct {
	timeout time.Duration
}

func NewMetadataFetcher() *MetadataFetcher {
	return &MetadataFetcher{
		timeout: 300 * time.Second,
	}
}

func (f *MetadataFetcher) SetTimeout(timeout time.Duration) {
	f.timeout = timeout
}

func (f *MetadataFetcher) FetchAlbum(spotifyURL string) (*AlbumMetadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), f.timeout)
	defer cancel()

	data, err := backend.GetFilteredSpotifyData(ctx, spotifyURL, false, 1*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Spotify metadata: %w", err)
	}

	albumPayload, ok := data.(*backend.AlbumResponsePayload)
	if !ok {
		return nil, fmt.Errorf("expected album data, got type %T (URL may not be an album)", data)
	}

	album := &AlbumMetadata{
		Name:        albumPayload.AlbumInfo.Name,
		Artist:      albumPayload.AlbumInfo.Artists,
		ReleaseDate: albumPayload.AlbumInfo.ReleaseDate,
		Images:      albumPayload.AlbumInfo.Images,
		TrackCount:  len(albumPayload.TrackList),
		Tracks:      make([]TrackMetadata, 0, len(albumPayload.TrackList)),
	}

	for _, track := range albumPayload.TrackList {
		album.Tracks = append(album.Tracks, TrackMetadata{
			ISRC:        track.ISRC,
			SpotifyID:   track.SpotifyID,
			Name:        track.Name,
			Artist:      track.Artists,
			AlbumArtist: track.AlbumArtist,
			AlbumName:   track.AlbumName,
			TrackNumber: track.TrackNumber,
			DiscNumber:  track.DiscNumber,
			TotalTracks: track.TotalTracks,
			Duration:    track.DurationMS,
			Images:      track.Images,
			ReleaseDate: track.ReleaseDate,
		})
	}

	return album, nil
}

func (f *MetadataFetcher) FetchPlaylist(spotifyURL string) (*PlaylistMetadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), f.timeout)
	defer cancel()

	data, err := backend.GetFilteredSpotifyData(ctx, spotifyURL, false, 1*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Spotify metadata: %w", err)
	}

	playlistPayload, ok := data.(backend.PlaylistResponsePayload)
	if !ok {
		return nil, fmt.Errorf("expected playlist data, got different type (URL may not be a playlist)")
	}

	playlist := &PlaylistMetadata{
		Name:       playlistPayload.PlaylistInfo.Owner.Name,
		Owner:      playlistPayload.PlaylistInfo.Owner.DisplayName,
		TrackCount: len(playlistPayload.TrackList),
		Tracks:     make([]TrackMetadata, 0, len(playlistPayload.TrackList)),
	}

	for _, track := range playlistPayload.TrackList {
		playlist.Tracks = append(playlist.Tracks, TrackMetadata{
			ISRC:        track.ISRC,
			SpotifyID:   track.SpotifyID,
			Name:        track.Name,
			Artist:      track.Artists,
			AlbumArtist: track.AlbumArtist,
			AlbumName:   track.AlbumName,
			TrackNumber: track.TrackNumber,
			DiscNumber:  track.DiscNumber,
			TotalTracks: track.TotalTracks,
			Duration:    track.DurationMS,
			Images:      track.Images,
			ReleaseDate: track.ReleaseDate,
		})
	}

	return playlist, nil
}

func (f *MetadataFetcher) FetchDiscography(spotifyURL string) (*DiscographyMetadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), f.timeout)
	defer cancel()

	data, err := backend.GetFilteredSpotifyData(ctx, spotifyURL, false, 1*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Spotify metadata: %w", err)
	}

	discographyPayload, ok := data.(*backend.ArtistDiscographyPayload)
	if !ok {
		return nil, fmt.Errorf("expected discography data, got type %T (URL may not be an artist discography)", data)
	}

	discography := &DiscographyMetadata{
		ArtistName:      discographyPayload.ArtistInfo.Name,
		DiscographyType: discographyPayload.ArtistInfo.DiscographyType,
		TotalAlbums:     len(discographyPayload.AlbumList),
		Albums:          make([]AlbumMetadata, 0, len(discographyPayload.AlbumList)),
		AllTracks:       make([]TrackMetadata, 0, len(discographyPayload.TrackList)),
	}

	tracksByAlbumID := make(map[string][]TrackMetadata)
	for _, track := range discographyPayload.TrackList {
		trackMeta := TrackMetadata{
			ISRC:        track.ISRC,
			SpotifyID:   track.SpotifyID,
			Name:        track.Name,
			Artist:      track.Artists,
			AlbumArtist: track.AlbumArtist,
			AlbumName:   track.AlbumName,
			TrackNumber: track.TrackNumber,
			DiscNumber:  track.DiscNumber,
			TotalTracks: track.TotalTracks,
			Duration:    track.DurationMS,
			Images:      track.Images,
			ReleaseDate: track.ReleaseDate,
		}
		discography.AllTracks = append(discography.AllTracks, trackMeta)
		tracksByAlbumID[track.AlbumName] = append(tracksByAlbumID[track.AlbumName], trackMeta)
	}

	for _, album := range discographyPayload.AlbumList {
		albumTracks := tracksByAlbumID[album.Name]
		artist := album.Artists
		if len(albumTracks) > 0 {
			artist = albumTracks[0].Artist
		}

		albumMeta := AlbumMetadata{
			Name:        album.Name,
			Artist:      artist,
			ReleaseDate: album.ReleaseDate,
			Images:      album.Images,
			TrackCount:  len(albumTracks),
			Tracks:      albumTracks,
		}
		discography.Albums = append(discography.Albums, albumMeta)
	}

	return discography, nil
}
