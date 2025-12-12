package hls

import (
	"bufio"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// Segment represents an HLS media segment.
type Segment struct {
	URL      string
	Sequence int
	Duration float64
}

// Playlist represents an HLS playlist with its segments.
type Playlist struct {
	Segments []*Segment
}

// ParsePlaylist parses an M3U8 playlist content and returns a list of segments.
// Uses pointers to reduce memory allocation overhead.
func ParsePlaylist(playlistContent, baseURL string) ([]*Segment, error) {
	var segments []*Segment
	var currentDuration float64
	var mediaSequence int

	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(playlistContent))

	// Pre-compiled regex patterns for better performance
	mediaSeqRegex := regexp.MustCompile(`#EXT-X-MEDIA-SEQUENCE:(\d+)`)
	durationRegex := regexp.MustCompile(`#EXTINF:([\d.]+)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if match := mediaSeqRegex.FindStringSubmatch(line); match != nil {
			mediaSequence, _ = strconv.Atoi(match[1])
			continue
		}

		if match := durationRegex.FindStringSubmatch(line); match != nil {
			currentDuration, _ = strconv.ParseFloat(match[1], 64)
			continue
		}

		// Segment URL line
		if line != "" && !strings.HasPrefix(line, "#") {
			segmentURL, err := base.Parse(line)
			if err != nil {
				return nil, fmt.Errorf("invalid segment URL %s: %w", line, err)
			}

			// Extract sequence number from segment URL if available
			seq := extractSequenceFromURL(line, mediaSequence)

			segments = append(segments, &Segment{
				URL:      segmentURL.String(),
				Sequence: seq,
				Duration: currentDuration,
			})

			mediaSequence++
			currentDuration = 0
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning playlist: %w", err)
	}

	return segments, nil
}

// GetLastSegment returns a pointer to the segment with the highest sequence number.
func GetLastSegment(segments []*Segment) *Segment {
	if len(segments) == 0 {
		return nil
	}

	// Find segment with highest sequence number
	last := segments[0]
	for i := 1; i < len(segments); i++ {
		if segments[i].Sequence > last.Sequence {
			last = segments[i]
		}
	}
	return last
}

// FindSegmentBySequence finds a segment by its sequence number.
// Returns a pointer to the segment if found, nil otherwise.
func FindSegmentBySequence(segments []*Segment, sequence int) *Segment {
	for _, seg := range segments {
		if seg.Sequence == sequence {
			return seg
		}
	}
	return nil
}

// extractSequenceFromURL extracts sequence number from segment URL.
// Expected format: master_1440_primary_719721.ts
func extractSequenceFromURL(segmentURL string, defaultSeq int) int {
	re := regexp.MustCompile(`_(\d+)\.ts`)
	matches := re.FindStringSubmatch(segmentURL)
	if len(matches) > 1 {
		if seq, err := strconv.Atoi(matches[1]); err == nil {
			return seq
		}
	}
	return defaultSeq
}
