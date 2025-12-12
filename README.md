# Stream Capture

A production-ready Go library and CLI tool for capturing HLS (HTTP Live Streaming) live streams by downloading and merging segments. Built with memory efficiency, extensibility, and ease of use in mind.

## 🚀 Features

### Core Functionality

- **Live Stream Capture**: Intelligent polling mechanism that continuously monitors M3U8 playlists and captures new segments as they become available in real-time
- **Memory Efficient**: Optimized for low memory usage through streaming I/O and pointer-based data structures, making it suitable for resource-constrained environments
- **Automatic Cleanup**: Temporary segments are stored in a temporary directory and automatically cleaned up after processing
- **Graceful Shutdown**: Supports context cancellation and signal handling (Ctrl+C) for clean termination
- **Thread-Safe Operations**: Concurrent segment tracking protected with mutexes ensures safe parallel operations

### Audio Processing

- **Audio Extraction**: Extract audio tracks from video streams using FFmpeg
- **MP3 Conversion**: Automatically converts extracted audio to MP3 format with high-quality encoding
- **Audio-Only Mode**: Option to extract only audio without saving the video file, saving disk space

### Subtitle Generation

- **AI-Powered Subtitles**: Leverages OpenAI Whisper for accurate speech-to-text conversion
- **Multiple Models**: Support for all Whisper models (tiny, base, small, medium, large, large-v2, large-v3) with configurable quality vs. speed trade-offs
- **Language Support**: Auto-detection or manual language specification for better accuracy
- **SRT Format**: Outputs industry-standard SubRip subtitle format compatible with all major video players

### Developer Experience

- **Clean Architecture**: Library-style structure following Go best practices, making it easy to integrate into other projects
- **Cobra CLI**: Modern, user-friendly command-line interface with comprehensive help and validation
- **Docker Support**: Fully containerized with multi-stage builds for minimal image size
- **CI/CD Ready**: GitHub Actions workflow for automated testing, building, and publishing

## 📦 Installation

### Using go install (Recommended)

Install the latest version:

```bash
go install github.com/bariiss/stream-capture/cmd/stream-capture@latest
```

Or install a specific version:

```bash
go install github.com/bariiss/stream-capture/cmd/stream-capture@v1.1.3
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

### Using Docker

#### Pull from GitHub Container Registry (Recommended)

```bash
docker pull ghcr.io/bariiss/stream-capture:latest
```

Or pull a specific version:

```bash
docker pull ghcr.io/bariiss/stream-capture:v1.1.3
```

#### Build locally

```bash
docker build -t bariiss/stream-capture .
```

#### Run directly

```bash
docker run --rm -v $(pwd)/output:/output \
  ghcr.io/bariiss/stream-capture:latest \
  -u "https://example.com/stream.m3u8" \
  -c 20 \
  -m /output/output.ts
```

### Using Docker Compose (Recommended for Whisper)

Docker Compose is the recommended way to use stream-capture with subtitle extraction, as it automatically manages Whisper model storage. Models are downloaded once and reused across runs, significantly speeding up subsequent subtitle extractions.

#### Option 1: Configure in docker-compose.yml

Edit `docker-compose.yml` and set your parameters:

```yaml
services:
  stream-capture:
    image: ghcr.io/bariiss/stream-capture:latest
    volumes:
      - ./output:/output
      - whisper-models:/root/.cache/whisper
    environment:
      - WHISPER_CACHE_DIR=/root/.cache/whisper
    command:
      - "-u"
      - "https://example.com/stream.m3u8"
      - "-c"
      - "10"
      - "--interval"
      - "2s"
      - "-m"
      - "/output/output.ts"
      - "--audio"
      - "--audio-output"
      - "/output/output.mp3"
      - "--subtitle"
      - "--subtitle-model"
      - "large"
      - "--subtitle-output"
      - "/output/output.srt"

volumes:
  whisper-models:
    driver: local
```

Then run:

```bash
# Create output directory
mkdir -p output

# Run with docker-compose (uses configuration from docker-compose.yml)
docker-compose up --rm
```

#### Option 2: Override command at runtime

Override parameters at runtime without modifying docker-compose.yml:

```bash
docker-compose run --rm stream-capture \
  -u "https://example.com/stream.m3u8" \
  -c 20 \
  -m /output/output.ts \
  --audio \
  --audio-output /output/output.mp3 \
  --subtitle \
  --subtitle-model large \
  --subtitle-output /output/output.srt
```

**Benefits of Docker Compose:**

- Persistent Whisper model storage: Models are cached in a volume and reused
- Faster subsequent runs: No need to re-download large Whisper models (e.g., large model is ~3GB)
- Consistent environment: Same dependencies across all machines
- Easy configuration: All settings in one file

## 📖 Usage

### Basic Command Structure

```bash
stream-capture -url <M3U8_URL> -count <SEGMENT_COUNT> -merge <OUTPUT_FILE> [OPTIONS]
# or
stream-capture -url <M3U8_URL> -count <SEGMENT_COUNT> -output <OUTPUT_FILE> [OPTIONS]
```

### Command-Line Parameters

#### Required Parameters

- `-u, --url <URL>`: M3U8 playlist URL (required)
  - The URL of the HLS playlist file (.m3u8)
  - Must be a valid HTTP/HTTPS URL

#### Output Parameters

- `-m, --merge <FILE>` or `-o, --output <FILE>`: Output file path for merged video segments
  - Required unless `--audio-only` is specified
  - Typically uses `.ts` extension for Transport Stream format
  - Alternative flags (`-m` and `-o`) provide the same functionality

#### Download Parameters

- `-c, --count <NUMBER>`: Number of segments to download (default: 10)
  - Starts from the latest available segment and downloads backwards
  - For live streams, the tool will wait for new segments if they're not immediately available
  - Higher values mean longer videos but more download time

- `-i, --interval <DURATION>`: Playlist polling interval (default: 2s)
  - How often to check the playlist for new segments
  - Format: `2s`, `500ms`, `3m`, etc.
  - Shorter intervals catch segments faster but use more bandwidth
  - Longer intervals save bandwidth but may miss segments in fast-changing streams

#### Audio Extraction Parameters

- `-a, --audio`: Extract audio as MP3 from the merged video file
  - Creates an MP3 file alongside the video file
  - Default output: `<video-file-name>.mp3` in the same directory
  - Requires FFmpeg to be installed

- `--audio-only`: Extract only audio without saving the video file
  - Video file is created temporarily for extraction, then deleted
  - Useful for audio-only use cases (podcasts, music streams)
  - Requires `--audio-output` to be specified

- `--audio-output <FILE>`: Custom output path for audio file
  - Required when using `--audio-only`
  - Optional when using `--audio` (defaults to `<video-file>.mp3`)
  - Should have `.mp3` extension

#### Subtitle Extraction Parameters

- `--subtitle`: Extract subtitles from audio using OpenAI Whisper
  - Automatically enables audio extraction (audio is needed for subtitle generation)
  - MP3 file is preserved after subtitle extraction (not deleted)
  - Requires Whisper to be installed
  - Output format: SRT (SubRip)

- `--subtitle-output <FILE>`: Custom output path for subtitle file
  - Optional: defaults to `<audio-file>.srt`
  - Should have `.srt` extension

- `--subtitle-language <CODE>`: Language code for subtitle extraction
  - Examples: `tr` (Turkish), `en` (English), `es` (Spanish), etc.
  - Optional: if not specified, Whisper will auto-detect the language
  - Specifying the language improves accuracy and speed

- `--subtitle-model <MODEL>`: Whisper model to use (default: `base`)
  - Available models: `tiny`, `base`, `small`, `medium`, `large`, `large-v2`, `large-v3`
  - **Speed vs. Accuracy Trade-off:**
    - `tiny`: Fastest, lowest accuracy (~39M parameters, ~75MB)
    - `base`: Good balance (default, ~74M parameters, ~142MB)
    - `small`: Better accuracy (~244M parameters, ~466MB)
    - `medium`: High accuracy (~769M parameters, ~1.5GB)
    - `large`: Best accuracy, slowest (~1550M parameters, ~3GB)
    - `large-v2`, `large-v3`: Latest large models with improvements

### Usage Examples

#### Basic Video Capture

```bash
# Download last 20 segments and merge into output.ts
stream-capture -u https://example.com/stream.m3u8 -c 20 -m output.ts
```

#### Video with Audio Extraction

```bash
# Download video and extract audio (creates both output.ts and output.mp3)
stream-capture -u https://example.com/stream.m3u8 -c 20 -m output.ts --audio

# Custom audio output path
stream-capture -u https://example.com/stream.m3u8 -c 20 -m output.ts --audio --audio-output music.mp3
```

#### Audio-Only Extraction

```bash
# Extract only audio (video file is deleted after extraction)
stream-capture -u https://example.com/stream.m3u8 -c 20 --audio-only --audio-output output.mp3
```

#### Subtitle Extraction

```bash
# Extract video, audio, and subtitles
stream-capture -u https://example.com/stream.m3u8 -c 20 -m output.ts --audio --subtitle

# With specific language (better accuracy)
stream-capture -u https://example.com/stream.m3u8 -c 20 -m output.ts --audio --subtitle --subtitle-language tr

# With high-accuracy model
stream-capture -u https://example.com/stream.m3u8 -c 20 -m output.ts --audio --subtitle --subtitle-model large

# Audio-only with subtitles (no video file saved)
stream-capture -u https://example.com/stream.m3u8 -c 20 --audio-only --audio-output output.mp3 --subtitle
```

#### Advanced Configuration

```bash
# Custom polling interval for fast-changing streams
stream-capture -u https://example.com/stream.m3u8 -c 30 -m video.ts -i 1s

# Full pipeline with all options
stream-capture \
  -u https://example.com/stream.m3u8 \
  -c 25 \
  -m video.ts \
  -i 2s \
  --audio \
  --audio-output audio.mp3 \
  --subtitle \
  --subtitle-model large \
  --subtitle-language en \
  --subtitle-output subtitles.srt
```

## 🔧 How It Works

### Capture Workflow

1. **Initial Playlist Fetch**: Downloads and parses the M3U8 playlist to identify available segments
2. **Segment Discovery**: Finds the last available segment sequence number
3. **Target Calculation**: Calculates the target segment range based on the count parameter
4. **Intelligent Polling**: For each segment:
   - Checks if the segment is available in the current playlist
   - If not available (common in live streams), polls the playlist at the specified interval
   - Waits until the segment becomes available or the stream ends
5. **Streaming Download**: Downloads segments directly to disk using streaming I/O to minimize memory usage
6. **Segment Merging**: Concatenates all downloaded segments into a single video file
7. **Post-Processing** (if requested):
   - Extracts audio using FFmpeg (if `--audio` or `--audio-only` is specified)
   - Generates subtitles using Whisper (if `--subtitle` is specified)
8. **Cleanup**: Automatically removes temporary files and directories

### Memory Efficiency

- **Streaming I/O**: Segments are streamed directly from HTTP responses to disk, never fully loaded into memory
- **Pointer Usage**: Extensive use of pointers reduces memory allocations and copying overhead
- **Temporary Storage**: Segments stored in temp directory are cleaned up immediately after merging
- **Minimal Memory Footprint**: Typically uses less than 50MB of RAM, even for large streams

### Live Stream Handling

For live streams where segments are constantly being added:

- The tool starts from the latest segment and works backwards
- If a required segment isn't available yet, it polls the playlist every `interval` seconds
- This ensures you capture the exact number of segments you requested, even if they're not all immediately available
- The tool continues until all requested segments are downloaded or the stream ends

## 🏗️ Architecture

### Project Structure

```text
stream-capture/
├── cmd/
│   └── stream-capture/          # CLI application entry point
│       ├── main.go              # Application entry point
│       └── cmd/
│           ├── root.go          # Cobra root command and flag definitions
│           └── capture.go       # Core capture logic and execution
├── internal/
│   ├── hls/                     # HLS playlist parsing and HTTP fetching
│   │   ├── playlist.go          # M3U8 playlist parsing logic
│   │   └── fetcher.go           # HTTP client for fetching playlists and segments
│   ├── downloader/              # Segment download and merging
│   │   └── manager.go           # Download coordination and segment management
│   ├── audio/                   # Audio extraction using FFmpeg
│   │   └── extractor.go         # FFmpeg audio extraction wrapper
│   └── subtitle/                # Subtitle generation using Whisper
│       └── extractor.go         # Whisper subtitle extraction wrapper
├── Dockerfile                   # Multi-stage Docker build
├── docker-compose.yml           # Docker Compose configuration
├── .github/
│   └── workflows/
│       ├── ci.yml               # CI workflow (testing, linting)
│       └── docker.yml           # Docker build and push workflow
├── go.mod                       # Go module definition
└── README.md                    # This file
```

### Internal Packages

#### `internal/hls`

Handles all HLS-related operations:

- **`ParsePlaylist()`**: Parses M3U8 playlists and extracts segment metadata
  - Supports `#EXTINF`, `#EXT-X-MEDIA-SEQUENCE`, and segment URL parsing
  - Handles both relative and absolute URLs
  - Returns structured segment information with sequence numbers and durations

- **`Fetcher`**: HTTP client for fetching playlists and segments
  - Configured with appropriate timeouts and user agent
  - Supports HTTP and HTTPS
  - Uses streaming for efficient memory usage

#### `internal/downloader`

Manages the download and merging process:

- **`Manager`**: Coordinates the entire download workflow
  - Creates temporary directory for segment storage
  - Tracks downloaded segments in a thread-safe map
  - Coordinates parallel downloads (future enhancement)
  - Merges segments using `cat` (POSIX) or `copy` (Windows) operations
  - Handles cleanup of temporary files

- **Thread Safety**: Uses `sync.RWMutex` to protect segment tracking map from concurrent access

#### `internal/audio`

FFmpeg integration for audio extraction:

- **`Extractor`**: Wraps FFmpeg for audio extraction
  - Detects FFmpeg installation in system PATH
  - Provides platform-specific installation hints if not found
  - Executes FFmpeg commands with appropriate encoding parameters
  - Supports MP3 encoding with high quality settings

#### `internal/subtitle`

OpenAI Whisper integration for subtitle generation:

- **`Extractor`**: Wraps Whisper CLI for subtitle extraction
  - Detects Whisper installation in system PATH
  - Provides platform-specific installation hints if not found
  - Supports all Whisper model sizes
  - Configurable language and output format
  - Handles SRT file generation and path management

### Design Decisions

#### Performance Optimizations

- **Streaming I/O**: Uses `io.Copy` to stream data directly from HTTP responses to files, avoiding large memory buffers
- **Pointer Usage**: Extensive use of pointers in data structures to minimize copying and reduce allocations
- **Temporary Files**: Segments stored in temp directory are cleaned up immediately after use

#### Reliability Features

- **Context Support**: All operations support context cancellation for graceful shutdown
- **Signal Handling**: Handles SIGINT and SIGTERM for clean termination
- **Error Handling**: Comprehensive error messages with context for easier debugging
- **Thread Safety**: Mutex-protected data structures ensure safe concurrent access

#### Extensibility

- **Library Structure**: Clean separation allows the tool to be used as a library in other projects
- **Modular Design**: Each package is independent and can be tested/modified separately
- **Plugin-Ready**: Architecture supports future additions (e.g., different subtitle providers, audio formats)

## 📋 Requirements

### Runtime Requirements

- **Go**: Version 1.25 or later (for building from source)
- **FFmpeg**: Required for audio extraction
  - **macOS**: `brew install ffmpeg`
  - **Ubuntu/Debian**: `apt-get install ffmpeg`
  - **Alpine**: `apk add ffmpeg`
  - **Windows**: Download from [ffmpeg.org](https://ffmpeg.org/download.html)
  - **Verification**: Run `ffmpeg -version` to verify installation

- **OpenAI Whisper**: Required for subtitle extraction
  - **macOS**: `brew install openai-whisper`
  - **Linux**: `pip install openai-whisper` (requires Python 3.8+)
  - **Windows**: `pip install openai-whisper` (requires Python 3.8+)
  - **Verification**: Run `whisper --help` to verify installation

### Docker Requirements

- **Docker**: Version 20.10 or later
- **Docker Compose**: Version 2.0 or later (for docker-compose.yml)
- The Docker image includes all dependencies (FFmpeg, Whisper, Python)

## 🐳 Docker Details

### Image Information

- **Registry**: GitHub Container Registry (GHCR)
- **Image**: `ghcr.io/bariiss/stream-capture`
- **Multi-Architecture**: Supports both `linux/amd64` and `linux/arm64`
- **Base Image**: Multi-stage build using `golang:1.25-alpine` (builder) and `python:3.12-slim` (runtime)
- **Size**: Optimized for minimal size while including all dependencies

### Build Arguments

The Dockerfile accepts the following build arguments:

- `GO_VERSION`: Go version for builder stage (default: 1.25)
- `VCS_REF`: Git commit SHA for labeling
- `BUILD_DATE`: Build timestamp for labeling
- `VERSION`: Version tag for labeling

### Labels

The image includes OCI labels for metadata:

- `org.opencontainers.image.title`
- `org.opencontainers.image.description`
- `org.opencontainers.image.version`
- `org.opencontainers.image.revision`
- `org.opencontainers.image.created`
- `org.opencontainers.image.source`

## 🔍 Troubleshooting

### Common Issues

#### FFmpeg Not Found

**Error**: `ffmpeg not found in PATH`

**Solution:**

1. Verify FFmpeg is installed: `ffmpeg -version`
2. Ensure FFmpeg is in your PATH
3. Check platform-specific installation instructions above

#### Whisper Not Found

**Error**: `whisper not found in PATH`

**Solution:**

1. Verify Whisper is installed: `whisper --help`
2. Ensure Whisper is in your PATH (usually `~/.local/bin` or `/usr/local/bin`)
3. On Linux/Windows, ensure Python and pip are installed first
4. Check platform-specific installation instructions above

#### Segments Not Downloading

**Symptoms**: Tool waits indefinitely for segments

**Possible Causes:**

1. Live stream ended before all segments were available
2. Network connectivity issues
3. Playlist URL is incorrect or expired
4. Segment URLs in playlist are invalid

**Solution:**

- Check the playlist URL manually with `curl` or browser
- Verify network connectivity
- Try reducing the segment count
- Increase polling interval if network is slow

#### Out of Memory

**Symptoms**: Process killed or crashes

**Solution:**

- The tool is designed to be memory-efficient, but very large streams may still cause issues
- Consider processing in smaller batches
- Ensure sufficient disk space for temporary files
- Check system memory limits

#### Docker Permission Issues

**Symptoms**: Cannot write to output directory

**Solution:**

```bash
# Ensure output directory has correct permissions
chmod 777 output/

# Or run with user mapping
docker run --rm -u $(id -u):$(id -g) -v $(pwd)/output:/output ...
```

### Performance Tips

1. **Subtitle Model Selection**: Use `tiny` or `base` for faster processing, `large` for better accuracy
2. **Polling Interval**: Adjust based on segment update frequency (shorter for fast streams, longer for slow ones)
3. **Docker Volume Caching**: Use Docker Compose to cache Whisper models (saves GBs of download time)
4. **Segment Count**: Only download as many segments as needed to minimize processing time

## 🔄 CI/CD

The project includes GitHub Actions workflows for automated testing and building:

- **`.github/workflows/ci.yml`**: Runs tests and linting on every push and PR
- **`.github/workflows/docker.yml`**: Builds and pushes Docker images on tag pushes

### Docker Image Publishing

Docker images are automatically built and pushed to GHCR when:

- A tag matching `v*` is pushed
- Multi-architecture builds (amd64, arm64)
- Tags include version tag, SHA tag, and `latest` (for main branch)

## 📝 License

[Your License Here]

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📧 Support

For issues, questions, or feature requests, please open an issue on GitHub.
