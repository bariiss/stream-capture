package subtitle

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Extractor handles subtitle extraction from audio files using OpenAI Whisper.
type Extractor struct {
	whisperPath string
}

// NewExtractor creates a new subtitle extractor with Whisper path detection.
func NewExtractor() (*Extractor, error) {
	// Try to find whisper in PATH
	whisperPath, err := exec.LookPath("whisper")
	if err != nil {
		installHint := getInstallHint()
		return nil, fmt.Errorf("whisper not found in PATH: %w\n%s", err, installHint)
	}

	return &Extractor{
		whisperPath: whisperPath,
	}, nil
}

// ExtractSubtitle extracts subtitles from an audio file using Whisper.
// Returns the path to the output subtitle file (SRT format).
func (e *Extractor) ExtractSubtitle(audioPath string, outputPath string, language string) error {
	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Whisper command arguments
	// --model: use base model (faster, good quality)
	// --output_dir: directory for output files
	// --output_format: srt format
	// --language: optional language code (e.g., "tr", "en")
	args := []string{
		audioPath,
		"--model", "base",
		"--output_dir", outputDir,
		"--output_format", "srt",
	}

	// Add language if specified
	if language != "" {
		args = append(args, "--language", language)
	}

	cmd := exec.Command(e.whisperPath, args...)

	// Capture both stdout and stderr for better error messages
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("whisper extraction failed: %w", err)
	}

	// Whisper creates output file with same name as input but with .srt extension
	// in the output directory. We need to check if our desired output path matches.
	audioBaseName := filepath.Base(audioPath)
	ext := filepath.Ext(audioBaseName)
	expectedSrtName := audioBaseName[:len(audioBaseName)-len(ext)] + ".srt"
	expectedSrtPath := filepath.Join(outputDir, expectedSrtName)

	// If the expected path doesn't match desired output path, rename it
	if expectedSrtPath != outputPath {
		if err := os.Rename(expectedSrtPath, outputPath); err != nil {
			// If rename fails, try to copy
			return fmt.Errorf("failed to move subtitle file to desired location: %w", err)
		}
	}

	return nil
}

// getInstallHint returns platform-specific installation instructions for Whisper.
func getInstallHint() string {
	switch runtime.GOOS {
	case "darwin":
		return "To install Whisper on macOS, run: brew install openai-whisper"
	case "linux":
		return "To install Whisper on Linux:\n" +
			"  Ubuntu/Debian: pip install openai-whisper (requires Python 3.8+)\n" +
			"  Or: sudo apt-get update && sudo apt-get install -y ffmpeg python3-pip && pip3 install openai-whisper\n" +
			"  Alpine: apk add py3-pip && pip install openai-whisper\n" +
			"  CentOS/RHEL: pip3 install openai-whisper (after installing Python 3.8+)"
	case "windows":
		return "To install Whisper on Windows:\n" +
			"  1. Install Python 3.8 or later from https://www.python.org/downloads/\n" +
			"  2. Open Command Prompt and run: pip install openai-whisper\n" +
			"  3. Make sure Python Scripts directory is in your PATH\n" +
			"  Or use pipx: pipx install openai-whisper"
	default:
		return "Please install Whisper for your platform. Visit https://github.com/openai/whisper\n" +
			"  Install with: pip install openai-whisper (requires Python 3.8+)"
	}
}
