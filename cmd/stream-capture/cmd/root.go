package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	playlistURL      string
	segmentCount     int
	mergeFile        string
	outputFile       string
	pollInterval     time.Duration
	extractAudio     bool
	audioOnly        bool
	audioOutput      string
	extractSubtitle  bool
	subtitleOutput   string
	subtitleLanguage string
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
	rootCmd.Flags().BoolVar(&audioOnly, "audio-only", false, "Extract only audio (video file will be deleted after extraction)")
	rootCmd.Flags().StringVar(&audioOutput, "audio-output", "", "Output path for audio file (default: <merge-file>.mp3)")
	rootCmd.Flags().BoolVar(&extractSubtitle, "subtitle", false, "Extract subtitles from audio using Whisper")
	rootCmd.Flags().StringVar(&subtitleOutput, "subtitle-output", "", "Output path for subtitle file (default: <audio-file>.srt)")
	rootCmd.Flags().StringVar(&subtitleLanguage, "subtitle-language", "", "Language code for subtitle extraction (e.g., tr, en). Auto-detect if not specified")
}

func runCapture(cmd *cobra.Command, args []string) error {
	// Use -merge if provided, otherwise use -output
	finalOutputFile := mergeFile
	if finalOutputFile == "" {
		finalOutputFile = outputFile
	}

	// If subtitle is enabled, automatically enable audio extraction (subtitle needs audio)
	if extractSubtitle {
		extractAudio = true
	}

	// If audio-only is enabled, automatically enable audio extraction
	if audioOnly {
		extractAudio = true
		// In audio-only mode, audio-output is required
		if audioOutput == "" {
			return fmt.Errorf("--audio-output is required when using --audio-only")
		}
		// If no output file specified, use temporary file
		if finalOutputFile == "" {
			finalOutputFile = os.TempDir() + "/stream-capture-temp.ts"
		}
	} else if finalOutputFile == "" {
		return fmt.Errorf("either -output or -merge flag is required")
	}

	// Import here to avoid circular dependencies
	return executeCapture(playlistURL, segmentCount, finalOutputFile, pollInterval, extractAudio, audioOnly, audioOutput, extractSubtitle, subtitleOutput, subtitleLanguage)
}
