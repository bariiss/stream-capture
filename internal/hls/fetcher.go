package hls

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Fetcher handles HTTP requests for HLS playlists and segments.
type Fetcher struct {
	client *http.Client
}

// NewFetcher creates a new Fetcher with default HTTP client.
func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchPlaylist fetches the M3U8 playlist from the given URL.
// Returns the playlist content as a string.
func (f *Fetcher) FetchPlaylist(url string) (string, error) {
	resp, err := f.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch playlist: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read playlist: %w", err)
	}

	return string(body), nil
}

// FetchSegment fetches a segment and writes it to the given writer.
// Uses streaming to reduce memory usage.
func (f *Fetcher) FetchSegment(segmentURL string, writer io.Writer) error {
	resp, err := f.client.Get(segmentURL)
	if err != nil {
		return fmt.Errorf("failed to fetch segment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write segment: %w", err)
	}

	return nil
}
