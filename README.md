# Stream Capture

A Go library and CLI tool for capturing HLS (HTTP Live Streaming) live streams by downloading and merging segments.

## Features

- **Live stream support**: Periodically polls the playlist and captures new segments as they become available
- **Memory efficient**: Uses streaming and pointer-based data structures to minimize memory usage
- **Temporary storage**: Segments are stored in a temporary directory and automatically cleaned up
- **Error handling**: Robust error handling with graceful shutdown support
- **Clean architecture**: Library-style structure following Go best practices

## Project Structure

```text
stream-capture/
├── cmd/
│   └── stream-capture/    # CLI application
├── internal/
│   ├── hls/               # HLS playlist parsing and fetching
│   └── downloader/        # Segment download and management
├── go.mod
└── README.md
```

## Installation

### Using go install (Recommended)

Install the latest version:

```bash
go install github.com/bariiss/stream-capture/cmd/stream-capture@latest
```

Or install a specific version:

```bash
go install github.com/bariiss/stream-capture/cmd/stream-capture@v1.0.3
```

After installation, make sure `$GOPATH/bin` or `$HOME/go/bin` is in your `PATH`:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Build from source

```bash
git clone https://github.com/bariiss/stream-capture.git
cd stream-capture
go build -o stream-capture ./cmd/stream-capture
```

Or use Docker:

```bash
docker build -t stream-capture .
docker run --rm -v $(pwd):/output stream-capture -url <M3U8_URL> -count 20 -output /output/output.ts
```

## Usage

```bash
stream-capture -url <M3U8_URL> -count <SEGMENT_COUNT> -merge <OUTPUT_FILE> [-interval <DURATION>] [-audio] [-audio-output <AUDIO_FILE>]
# or
stream-capture -url <M3U8_URL> -count <SEGMENT_COUNT> -output <OUTPUT_FILE> [-interval <DURATION>] [-audio] [-audio-output <AUDIO_FILE>]
```

### Parameters

- `-url` (required): M3U8 playlist URL
- `-count` (default: 10): Number of segments to download, starting from the latest available segment
- `-merge` or `-output` (required): Output file path for merged segments
- `-interval` (default: 2s): Playlist polling interval
- `-audio` (optional): Extract audio as MP3 from the merged video file
- `-audio-output` (optional): Output path for audio file (default: `<merge-file>.mp3`)

### Examples

```bash
# Download last 20 segments and merge into output.ts
./stream-capture -url https://example.com/stream.m3u8 -count 20 -merge output.ts

# Download and extract audio as MP3 (outputs to output.mp3)
./stream-capture -url https://example.com/stream.m3u8 -count 20 -merge output.ts -audio

# Download with custom audio output path
./stream-capture -url https://example.com/stream.m3u8 -count 20 -merge output.ts -audio -audio-output music.mp3

# Custom polling interval
./stream-capture -url https://example.com/stream.m3u8 -count 30 -merge video.ts -interval 3s
```

## How It Works

1. Fetches the initial M3U8 playlist and finds the last available segment
2. Starts downloading from the last segment onwards
3. For each segment:
   - Polls the playlist until the segment becomes available
   - Downloads the segment to a temporary directory using streaming
4. Merges all downloaded segments into the output file
5. Cleans up the temporary directory automatically

## Architecture

### Internal Packages

- **`internal/hls`**: Handles HLS playlist parsing and HTTP fetching
  - `ParsePlaylist()`: Parses M3U8 playlists and extracts segments
  - `Fetcher`: HTTP client for fetching playlists and segments

- **`internal/downloader`**: Manages segment downloading and merging
  - `Manager`: Coordinates downloads, temporary storage, and merging
  - Uses streaming to minimize memory footprint
  - Automatic cleanup of temporary files

- **`internal/audio`**: Handles audio extraction from video files
  - `Extractor`: Uses FFmpeg to extract audio and convert to MP3
  - Supports MP3 encoding with configurable bitrate and sample rate
  - Requires FFmpeg to be installed in the system

### Design Decisions

- **Pointers**: Used extensively to reduce memory allocations and copying
- **Streaming**: Segments are streamed directly from HTTP response to disk
- **Temporary storage**: Segments stored in temp directory, cleaned up after merge
- **Context support**: Graceful shutdown via context cancellation
- **Thread-safe**: Concurrent access to segment maps protected with mutexes

## Requirements

- Go 1.25 or later
- FFmpeg (for audio extraction feature)
  - On macOS: `brew install ffmpeg`
  - On Ubuntu/Debian: `apt-get install ffmpeg`
  - On Alpine: `apk add ffmpeg`
  - On Windows: Download from [ffmpeg.org](https://ffmpeg.org/download.html)

## License

[Your License Here]
