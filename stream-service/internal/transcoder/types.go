package transcoder

import (
	"fmt"
	"strconv"
	"strings"
)

// ABRVariant represents a single quality variant for adaptive bitrate streaming
type ABRVariant struct {
	Name         string // e.g., "1080p", "720p", "480p", "360p"
	Resolution   string // e.g., "1920x1080"
	Width        int
	Height       int
	VideoBitrate string // e.g., "5000k"
	AudioBitrate string // e.g., "128k"
	Bandwidth    int    // Total bandwidth in bits per second
}

// ABRConfig holds configuration for adaptive bitrate streaming
type ABRConfig struct {
	Enabled  bool
	Variants []ABRVariant
}

// DefaultABRVariants returns the default set of ABR variants
func DefaultABRVariants() []ABRVariant {
	return []ABRVariant{
		{
			Name:         "1080p",
			Resolution:   "1920x1080",
			Width:        1920,
			Height:       1080,
			VideoBitrate: "5000k",
			AudioBitrate: "128k",
			Bandwidth:    5128000, // 5000k video + 128k audio
		},
		{
			Name:         "720p",
			Resolution:   "1280x720",
			Width:        1280,
			Height:       720,
			VideoBitrate: "2500k",
			AudioBitrate: "128k",
			Bandwidth:    2628000, // 2500k video + 128k audio
		},
		{
			Name:         "480p",
			Resolution:   "854x480",
			Width:        854,
			Height:       480,
			VideoBitrate: "1000k",
			AudioBitrate: "96k",
			Bandwidth:    1096000, // 1000k video + 96k audio
		},
		{
			Name:         "360p",
			Resolution:   "640x360",
			Width:        640,
			Height:       360,
			VideoBitrate: "600k",
			AudioBitrate: "64k",
			Bandwidth:    664000, // 600k video + 64k audio
		},
	}
}

// ParseBitrate converts a bitrate string (e.g., "5000k") to bits per second
func ParseBitrate(bitrate string) (int, error) {
	bitrate = strings.ToLower(strings.TrimSpace(bitrate))

	if strings.HasSuffix(bitrate, "k") {
		val, err := strconv.Atoi(strings.TrimSuffix(bitrate, "k"))
		if err != nil {
			return 0, fmt.Errorf("invalid bitrate format: %s", bitrate)
		}
		return val * 1000, nil
	}

	if strings.HasSuffix(bitrate, "m") {
		val, err := strconv.Atoi(strings.TrimSuffix(bitrate, "m"))
		if err != nil {
			return 0, fmt.Errorf("invalid bitrate format: %s", bitrate)
		}
		return val * 1000000, nil
	}

	// Assume it's already in bps if no suffix
	val, err := strconv.Atoi(bitrate)
	if err != nil {
		return 0, fmt.Errorf("invalid bitrate format: %s", bitrate)
	}
	return val, nil
}

// CalculateMaxRate calculates the maxrate for CBR encoding (typically 1.5x bitrate)
func CalculateMaxRate(bitrate string) string {
	bps, err := ParseBitrate(bitrate)
	if err != nil {
		return bitrate // Return original if parsing fails
	}

	maxRate := int(float64(bps) * 1.5)
	return fmt.Sprintf("%dk", maxRate/1000)
}

// CalculateBufSize calculates the buffer size for CBR encoding (typically 2x bitrate)
func CalculateBufSize(bitrate string) string {
	bps, err := ParseBitrate(bitrate)
	if err != nil {
		return bitrate // Return original if parsing fails
	}

	bufSize := bps * 2
	return fmt.Sprintf("%dk", bufSize/1000)
}

// ParseResolution parses a resolution string like "1920x1080" into width and height
func ParseResolution(resolution string) (width int, height int, err error) {
	parts := strings.Split(resolution, "x")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid resolution format: %s", resolution)
	}

	width, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid width in resolution: %s", resolution)
	}

	height, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid height in resolution: %s", resolution)
	}

	return width, height, nil
}
