package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Extractor handles audio extraction from video files using FFmpeg.
type Extractor struct {
	ffmpegPath string
}

// NewExtractor creates a new audio extractor with FFmpeg path detection.
func NewExtractor() (*Extractor, error) {
	// Try to find ffmpeg in PATH
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		installHint := getInstallHint()
		return nil, fmt.Errorf("ffmpeg not found in PATH: %w\n%s", err, installHint)
	}

	return &Extractor{
		ffmpegPath: ffmpegPath,
	}, nil
}

// ExtractAudio extracts audio from a video file and saves it as MP3.
// Returns the path to the output MP3 file.
func (e *Extractor) ExtractAudio(videoPath string, outputPath string) error {
	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// FFmpeg command to extract audio and convert to MP3
	// -i: input file
	// -vn: no video
	// -acodec libmp3lame: use MP3 codec
	// -ab 192k: audio bitrate 192kbps
	// -ar 44100: audio sample rate 44.1kHz
	// -y: overwrite output file if exists
	cmd := exec.Command(e.ffmpegPath,
		"-i", videoPath,
		"-vn",
		"-acodec", "libmp3lame",
		"-ab", "192k",
		"-ar", "44100",
		"-y",
		outputPath,
	)

	// Capture both stdout and stderr for better error messages
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg extraction failed: %w", err)
	}

	return nil
}

// ExtractAudioFromTS extracts audio from a TS (Transport Stream) file.
// This is a convenience method specifically for HLS segment files.
func (e *Extractor) ExtractAudioFromTS(tsPath string, outputPath string) error {
	return e.ExtractAudio(tsPath, outputPath)
}

// getInstallHint returns platform-specific installation instructions for FFmpeg.
func getInstallHint() string {
	switch runtime.GOOS {
	case "darwin":
		return "To install FFmpeg on macOS, run: brew install ffmpeg"
	case "linux":
		return "To install FFmpeg on Linux:\n" +
			"  Ubuntu/Debian: sudo apt-get update && sudo apt-get install -y ffmpeg\n" +
			"  Alpine: sudo apk add ffmpeg\n" +
			"  CentOS/RHEL: sudo yum install ffmpeg (or sudo dnf install ffmpeg)"
	case "windows":
		return "To install FFmpeg on Windows:\n" +
			"  1. Download from https://ffmpeg.org/download.html\n" +
			"  2. Extract and add the bin directory to your PATH environment variable\n" +
			"  Or use Chocolatey: choco install ffmpeg\n" +
			"  Or use Scoop: scoop install ffmpeg"
	default:
		return "Please install FFmpeg for your platform. Visit https://ffmpeg.org/download.html"
	}
}
