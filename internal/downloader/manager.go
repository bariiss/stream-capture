package downloader

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/bariiss/stream-capture/internal/hls"
)

// Manager handles downloading and managing HLS segments.
type Manager struct {
	fetcher  *hls.Fetcher
	tempDir  string
	segments map[int]string // sequence -> file path
	mu       sync.RWMutex
}

// NewManager creates a new download manager with a temporary directory.
func NewManager(tempDir string) (*Manager, error) {
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return &Manager{
		fetcher:  hls.NewFetcher(),
		tempDir:  tempDir,
		segments: make(map[int]string),
	}, nil
}

// DownloadSegment downloads a segment to the temporary directory.
// Returns the file path if successful.
func (m *Manager) DownloadSegment(segment *hls.Segment) (string, error) {
	m.mu.RLock()
	if path, exists := m.segments[segment.Sequence]; exists {
		// Check if file still exists
		if _, err := os.Stat(path); err == nil {
			m.mu.RUnlock()
			return path, nil
		}
		// File doesn't exist, remove from map
		delete(m.segments, segment.Sequence)
	}
	m.mu.RUnlock()

	// Create segment file
	filename := filepath.Join(m.tempDir, fmt.Sprintf("segment_%d.ts", segment.Sequence))
	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create segment file: %w", err)
	}
	defer file.Close()

	// Download segment using streaming to reduce memory usage
	if err := m.fetcher.FetchSegment(segment.URL, file); err != nil {
		os.Remove(filename) // Clean up on error
		return "", err
	}

	// Store in map
	m.mu.Lock()
	m.segments[segment.Sequence] = filename
	m.mu.Unlock()

	return filename, nil
}

// GetSegmentPath returns the file path for a given sequence number.
func (m *Manager) GetSegmentPath(sequence int) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	path, exists := m.segments[sequence]
	return path, exists
}

// MergeSegments merges all downloaded segments into a single output file.
// Uses streaming to reduce memory usage.
func (m *Manager) MergeSegments(outputPath string, sequences []int) error {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	for _, seq := range sequences {
		m.mu.RLock()
		segmentPath, exists := m.segments[seq]
		m.mu.RUnlock()

		if !exists {
			return fmt.Errorf("segment %d not found", seq)
		}

		if err := copyFile(segmentPath, outputFile); err != nil {
			return fmt.Errorf("failed to copy segment %d: %w", seq, err)
		}
	}

	return nil
}

// Cleanup removes all downloaded segments and the temporary directory.
func (m *Manager) Cleanup() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, path := range m.segments {
		os.Remove(path)
	}
	m.segments = make(map[int]string)

	return os.RemoveAll(m.tempDir)
}

// copyFile copies a file to a writer using streaming.
func copyFile(srcPath string, dst io.Writer) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	_, err = io.Copy(dst, src)
	return err
}
