package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	playlistURL  string
	segmentCount int
	mergeFile    string
	outputFile   string
	pollInterval time.Duration
	extractAudio bool
	audioOutput  string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "stream-capture",
	Short: "HLS stream capture tool",
	Long: `A Go library and CLI tool for capturing HLS (HTTP Live Streaming) live streams
by downloading and merging segments. Supports audio extraction using FFmpeg.`,
	RunE: runCapture,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Required flags
	rootCmd.Flags().StringVarP(&playlistURL, "url", "u", "", "M3U8 playlist URL (required)")
	rootCmd.MarkFlagRequired("url")

	// Optional flags
	rootCmd.Flags().IntVarP(&segmentCount, "count", "c", 10, "Number of segments to download (starting from the latest)")
	rootCmd.Flags().StringVarP(&mergeFile, "merge", "m", "", "Output file for merged segments (alternative to -output)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for merged segments (alternative to -merge)")
	rootCmd.Flags().DurationVarP(&pollInterval, "interval", "i", 2*time.Second, "Playlist polling interval")
	rootCmd.Flags().BoolVarP(&extractAudio, "audio", "a", false, "Extract audio as MP3 from the merged video file")
	rootCmd.Flags().StringVar(&audioOutput, "audio-output", "", "Output path for audio file (default: <merge-file>.mp3)")
}

func runCapture(cmd *cobra.Command, args []string) error {
	// Use -merge if provided, otherwise use -output
	finalOutputFile := mergeFile
	if finalOutputFile == "" {
		finalOutputFile = outputFile
	}

	if finalOutputFile == "" {
		return fmt.Errorf("either -output or -merge flag is required")
	}

	// Import here to avoid circular dependencies
	return executeCapture(playlistURL, segmentCount, finalOutputFile, pollInterval, extractAudio, audioOutput)
}
