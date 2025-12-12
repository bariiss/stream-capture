package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/bariiss/stream-capture/internal/downloader"
	"github.com/bariiss/stream-capture/internal/hls"
)

func main() {
	var (
		playlistURL  = flag.String("url", "", "M3U8 playlist URL")
		segmentCount = flag.Int("count", 10, "Number of segments to download (starting from the latest)")
		outputFile   = flag.String("output", "", "Output file for merged segments (alternative to -merge)")
		mergeFile    = flag.String("merge", "", "Output file for merged segments (alternative to -output)")
		pollInterval = flag.Duration("interval", 2*time.Second, "Playlist polling interval")
	)
	flag.Parse()

	if *playlistURL == "" {
		fmt.Fprintf(os.Stderr, "Error: -url flag is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Use -merge if provided, otherwise use -output
	finalOutputFile := *mergeFile
	if finalOutputFile == "" {
		finalOutputFile = *outputFile
	}

	if finalOutputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: either -output or -merge flag is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Create temporary directory for segments
	tempDir, err := os.MkdirTemp("", "stream-capture-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	// Create download manager
	manager, err := downloader.NewManager(tempDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating download manager: %v\n", err)
		os.Exit(1)
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
	fmt.Printf("Playlist URL: %s\n", *playlistURL)
	fmt.Printf("Target segments: %d\n", *segmentCount)
	fmt.Printf("Polling interval: %v\n", *pollInterval)
	fmt.Printf("Temp directory: %s\n\n", tempDir)

	// Create HLS fetcher
	fetcher := hls.NewFetcher()

	// Fetch initial playlist
	playlistContent, err := fetcher.FetchPlaylist(*playlistURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching playlist: %v\n", err)
		os.Exit(1)
	}

	segments, err := hls.ParsePlaylist(playlistContent, *playlistURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing playlist: %v\n", err)
		os.Exit(1)
	}

	if len(segments) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No segments found in playlist\n")
		os.Exit(1)
	}

	// Find last segment
	lastSegment := hls.GetLastSegment(segments)
	if lastSegment == nil {
		fmt.Fprintf(os.Stderr, "Error: Could not determine last segment\n")
		os.Exit(1)
	}

	startSequence := lastSegment.Sequence
	targetSequence := startSequence + *segmentCount - 1

	fmt.Printf("Starting from segment %d, target: %d (need %d segments)\n\n", startSequence, targetSequence, *segmentCount)

	// Download segments
	downloadedSequences := make([]int, 0, *segmentCount)
	for currentSeq := startSequence; currentSeq <= targetSequence; currentSeq++ {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			fmt.Println("Cancelled by user")
			return
		default:
		}

		// Wait for segment to be available
		var segment *hls.Segment
		retryCount := 0
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Cancelled by user")
				return
			default:
			}

			playlistContent, err := fetcher.FetchPlaylist(*playlistURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error fetching playlist: %v\n", err)
				time.Sleep(*pollInterval)
				continue
			}

			segments, err := hls.ParsePlaylist(playlistContent, *playlistURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing playlist: %v\n", err)
				time.Sleep(*pollInterval)
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
			time.Sleep(*pollInterval)
		}

		// Download segment
		fmt.Printf("[%d/%d] Downloading segment %d: %s\n", currentSeq-startSequence+1, *segmentCount, currentSeq, filepath.Base(segment.URL))

		_, err := manager.DownloadSegment(segment)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error downloading segment %d: %v\n", currentSeq, err)
			continue
		}

		downloadedSequences = append(downloadedSequences, currentSeq)
	}

	fmt.Printf("\nSuccessfully downloaded %d segments\n", len(downloadedSequences))

	// Merge segments
	fmt.Printf("Merging segments into: %s\n", finalOutputFile)

	// Ensure output directory exists
	outputDir := filepath.Dir(finalOutputFile)
	if outputDir != "" && outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
			os.Exit(1)
		}
	}

	if err := manager.MergeSegments(finalOutputFile, downloadedSequences); err != nil {
		fmt.Fprintf(os.Stderr, "Error merging segments: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully merged segments into %s\n", finalOutputFile)
	fmt.Println("Temp directory cleaned up")
}
