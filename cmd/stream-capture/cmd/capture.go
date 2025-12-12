package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/bariiss/stream-capture/internal/audio"
	"github.com/bariiss/stream-capture/internal/downloader"
	"github.com/bariiss/stream-capture/internal/hls"
	"github.com/bariiss/stream-capture/internal/subtitle"
)

// executeCapture performs the actual stream capture process
func executeCapture(
	playlistURL string,
	segmentCount int,
	outputFile string,
	pollInterval time.Duration,
	extractAudio bool,
	audioOnly bool,
	audioOutput string,
	extractSubtitle bool,
	subtitleOutput string,
	subtitleLanguage string,
	subtitleModel string,
) error {
	// Create temporary directory for segments
	tempDir, err := os.MkdirTemp("", "stream-capture-*")
	if err != nil {
		return fmt.Errorf("error creating temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create download manager
	manager, err := downloader.NewManager(tempDir)
	if err != nil {
		return fmt.Errorf("error creating download manager: %w", err)
	}
	defer manager.Cleanup()

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		cancel()
	}()

	fmt.Printf("Live stream capture started\n")
	fmt.Printf("Playlist URL: %s\n", playlistURL)
	fmt.Printf("Target segments: %d\n", segmentCount)
	fmt.Printf("Polling interval: %v\n", pollInterval)
	fmt.Printf("Temp directory: %s\n\n", tempDir)

	// Create HLS fetcher
	fetcher := hls.NewFetcher()

	// Fetch initial playlist
	playlistContent, err := fetcher.FetchPlaylist(playlistURL)
	if err != nil {
		return fmt.Errorf("error fetching playlist: %w", err)
	}

	segments, err := hls.ParsePlaylist(playlistContent, playlistURL)
	if err != nil {
		return fmt.Errorf("error parsing playlist: %w", err)
	}

	if len(segments) == 0 {
		return fmt.Errorf("no segments found in playlist")
	}

	// Find last segment
	lastSegment := hls.GetLastSegment(segments)
	if lastSegment == nil {
		return fmt.Errorf("could not determine last segment")
	}

	startSequence := lastSegment.Sequence
	targetSequence := startSequence + segmentCount - 1

	fmt.Printf("Starting from segment %d, target: %d (need %d segments)\n\n", startSequence, targetSequence, segmentCount)

	// Download segments
	downloadedSequences := make([]int, 0, segmentCount)
	for currentSeq := startSequence; currentSeq <= targetSequence; currentSeq++ {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			fmt.Println("Cancelled by user")
			return nil
		default:
		}

		// Wait for segment to be available
		var segment *hls.Segment
		retryCount := 0
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Cancelled by user")
				return nil
			default:
			}

			playlistContent, err := fetcher.FetchPlaylist(playlistURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error fetching playlist: %v\n", err)
				time.Sleep(pollInterval)
				continue
			}

			segments, err := hls.ParsePlaylist(playlistContent, playlistURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing playlist: %v\n", err)
				time.Sleep(pollInterval)
				continue
			}

			segment = hls.FindSegmentBySequence(segments, currentSeq)
			if segment != nil {
				break
			}

			lastSeg := hls.GetLastSegment(segments)
			if retryCount%5 == 0 || retryCount == 0 {
				fmt.Printf("Waiting for segment %d... (current last: %d)\n", currentSeq, lastSeg.Sequence)
			}
			retryCount++
			time.Sleep(pollInterval)
		}

		// Download segment
		fmt.Printf("[%d/%d] Downloading segment %d: %s\n", currentSeq-startSequence+1, segmentCount, currentSeq, filepath.Base(segment.URL))

		_, err := manager.DownloadSegment(segment)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error downloading segment %d: %v\n", currentSeq, err)
			continue
		}

		downloadedSequences = append(downloadedSequences, currentSeq)
	}

	fmt.Printf("\nSuccessfully downloaded %d segments\n", len(downloadedSequences))

	// Merge segments
	fmt.Printf("Merging segments into: %s\n", outputFile)

	// Ensure output directory exists
	outputDir := filepath.Dir(outputFile)
	if outputDir != "" && outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("error creating output directory: %w", err)
		}
	}

	// Only merge video if not audio-only mode
	var tempVideoFile string
	if !audioOnly {
		if err := manager.MergeSegments(outputFile, downloadedSequences); err != nil {
			return fmt.Errorf("error merging segments: %w", err)
		}
		fmt.Printf("Successfully merged segments into %s\n", outputFile)
		tempVideoFile = outputFile
	} else {
		// For audio-only, create temporary video file
		tempVideoFile = outputFile
		if err := manager.MergeSegments(tempVideoFile, downloadedSequences); err != nil {
			return fmt.Errorf("error merging segments: %w", err)
		}
		fmt.Printf("Merged segments to temporary file for audio extraction\n")
	}

	// Extract audio if requested
	if extractAudio {
		audioExtractor, err := audio.NewExtractor()
		if err != nil {
			return fmt.Errorf("error initializing audio extractor: %w", err)
		}

		// Determine audio output path
		audioOutputPath := audioOutput
		if audioOutputPath == "" {
			// Default to same name as video file but with .mp3 extension
			ext := filepath.Ext(outputFile)
			audioOutputPath = outputFile[:len(outputFile)-len(ext)] + ".mp3"
		}

		fmt.Printf("Extracting audio to: %s\n", audioOutputPath)
		if err := audioExtractor.ExtractAudio(tempVideoFile, audioOutputPath); err != nil {
			return fmt.Errorf("error extracting audio: %w", err)
		}
		fmt.Printf("Successfully extracted audio to %s\n", audioOutputPath)

		// Extract subtitles if requested
		if extractSubtitle {
			subtitleExtractor, err := subtitle.NewExtractor()
			if err != nil {
				return fmt.Errorf("error initializing subtitle extractor: %w", err)
			}

			// Determine subtitle output path
			subtitleOutputPath := subtitleOutput
			if subtitleOutputPath == "" {
				// Default to same name as audio file but with .srt extension
				ext := filepath.Ext(audioOutputPath)
				subtitleOutputPath = audioOutputPath[:len(audioOutputPath)-len(ext)] + ".srt"
			}

			fmt.Printf("Extracting subtitles to: %s (model: %s)\n", subtitleOutputPath, subtitleModel)
			if err := subtitleExtractor.ExtractSubtitle(audioOutputPath, subtitleOutputPath, subtitleLanguage, subtitleModel); err != nil {
				return fmt.Errorf("error extracting subtitles: %w", err)
			}
			fmt.Printf("Successfully extracted subtitles to %s\n", subtitleOutputPath)
		}

		// If audio-only mode, delete the video file
		if audioOnly {
			if err := os.Remove(tempVideoFile); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to remove temporary video file: %v\n", err)
			} else {
				fmt.Printf("Removed temporary video file: %s\n", tempVideoFile)
			}
		}
	}

	fmt.Println("Temp directory cleaned up")
	return nil
}
