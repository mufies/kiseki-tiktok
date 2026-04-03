package transcoder

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kiseki/stream-service/config"
)

// HLSTranscoder manages FFmpeg process for converting RTMP to HLS
type HLSTranscoder struct {
	streamID     uuid.UUID
	streamKey    string
	config       *config.Config
	abrConfig    *ABRConfig
	cmd          *exec.Cmd
	stdinPipe    io.WriteCloser
	flvWriter    *FLVWriter
	hlsOutputDir string
	isRunning    bool
	mu           sync.RWMutex

	// Statistics
	startTime    time.Time
	bytesWritten int64
}

// NewHLSTranscoder creates a new HLS transcoder for a stream
func NewHLSTranscoder(streamID uuid.UUID, streamKey string, cfg *config.Config) (*HLSTranscoder, error) {
	// Create output directory for HLS files
	// Format: /tmp/hls/{stream_id}/
	hlsOutputDir := filepath.Join("/tmp", "hls", streamID.String())
	if err := os.MkdirAll(hlsOutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create HLS output directory: %w", err)
	}

	// Build ABR configuration
	abrConfig := buildABRConfig(cfg)

	// Create variant subdirectories if ABR is enabled
	if abrConfig.Enabled {
		for _, variant := range abrConfig.Variants {
			variantDir := filepath.Join(hlsOutputDir, variant.Name)
			if err := os.MkdirAll(variantDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create variant directory %s: %w", variant.Name, err)
			}
		}
	}

	return &HLSTranscoder{
		streamID:     streamID,
		streamKey:    streamKey,
		config:       cfg,
		abrConfig:    abrConfig,
		hlsOutputDir: hlsOutputDir,
		isRunning:    false,
	}, nil
}

// Start starts the FFmpeg transcoding process
func (t *HLSTranscoder) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isRunning {
		return fmt.Errorf("transcoder already running")
	}

	// Build FFmpeg arguments based on ABR configuration
	args := t.buildFFmpegArgs()

	log.Printf("[Transcoder] Starting FFmpeg for stream %s", t.streamID)
	log.Printf("[Transcoder] Output directory: %s", t.hlsOutputDir)
	log.Printf("[Transcoder] ABR enabled: %v", t.abrConfig.Enabled)
	if t.abrConfig.Enabled {
		log.Printf("[Transcoder] Variants: %d", len(t.abrConfig.Variants))
		for _, v := range t.abrConfig.Variants {
			log.Printf("[Transcoder]   - %s: %dx%d @ %s video, %s audio",
				v.Name, v.Width, v.Height, v.VideoBitrate, v.AudioBitrate)
		}
	}
	log.Printf("[Transcoder] Command: %s %v", t.config.FFmpegPath, args)

	t.cmd = exec.Command(t.config.FFmpegPath, args...)

	// Capture FFmpeg stderr for debugging
	t.cmd.Stderr = os.Stderr
	t.cmd.Stdout = os.Stdout

	// Get stdin pipe to write RTMP data
	var err error
	t.stdinPipe, err = t.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	// Start FFmpeg process
	if err := t.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start FFmpeg: %w", err)
	}

	// Create FLV writer for this transcoder
	t.flvWriter = NewFLVWriter(t.stdinPipe)

	t.isRunning = true
	t.startTime = time.Now()

	log.Printf("[Transcoder] ✅ FFmpeg started for stream %s (PID: %d)", t.streamID, t.cmd.Process.Pid)

	// Generate master playlist for ABR
	if t.abrConfig.Enabled && len(t.abrConfig.Variants) > 1 {
		if err := t.GenerateMasterPlaylist(); err != nil {
			log.Printf("[Transcoder] ⚠️  Failed to generate master playlist: %v", err)
			// Continue anyway - playlist will be generated or updated later
		}
	}

	// Monitor FFmpeg process
	go t.monitorProcess()

	return nil
}

// WriteData writes RTMP stream data to FFmpeg stdin
func (t *HLSTranscoder) WriteData(data []byte) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.isRunning || t.stdinPipe == nil {
		return fmt.Errorf("transcoder not running")
	}

	n, err := t.stdinPipe.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to FFmpeg: %w", err)
	}

	t.bytesWritten += int64(n)
	return nil
}

// Stop stops the FFmpeg process
func (t *HLSTranscoder) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.isRunning {
		return nil
	}

	log.Printf("[Transcoder] Stopping FFmpeg for stream %s", t.streamID)

	// Close stdin to signal EOF to FFmpeg
	if t.stdinPipe != nil {
		if err := t.stdinPipe.Close(); err != nil {
			log.Printf("[Transcoder] Error closing stdin: %v", err)
		}
	}

	// Wait for FFmpeg to finish processing
	if t.cmd != nil && t.cmd.Process != nil {
		// Give FFmpeg some time to finish gracefully
		done := make(chan error, 1)
		go func() {
			done <- t.cmd.Wait()
		}()

		select {
		case err := <-done:
			if err != nil {
				log.Printf("[Transcoder] FFmpeg exited with error: %v", err)
			}
		case <-time.After(5 * time.Second):
			// Force kill if it doesn't stop gracefully
			log.Printf("[Transcoder] FFmpeg didn't stop gracefully, killing process")
			if err := t.cmd.Process.Kill(); err != nil {
				log.Printf("[Transcoder] Error killing FFmpeg: %v", err)
			}
		}
	}

	duration := time.Since(t.startTime)
	log.Printf("[Transcoder] ✅ Stopped FFmpeg for stream %s", t.streamID)
	log.Printf("[Transcoder]    Duration: %s", duration.Round(time.Second))
	log.Printf("[Transcoder]    Bytes written: %.2f MB", float64(t.bytesWritten)/(1024*1024))

	t.isRunning = false

	// Optional: Clean up HLS files after stream ends
	// Uncomment if you want to delete files immediately
	// defer os.RemoveAll(t.hlsOutputDir)

	return nil
}

// monitorProcess monitors the FFmpeg process and logs when it exits
func (t *HLSTranscoder) monitorProcess() {
	err := t.cmd.Wait()

	t.mu.Lock()
	t.isRunning = false
	t.mu.Unlock()

	if err != nil {
		log.Printf("[Transcoder] ❌ FFmpeg process exited with error for stream %s: %v", t.streamID, err)
	} else {
		log.Printf("[Transcoder] FFmpeg process exited normally for stream %s", t.streamID)
	}
}

// GetPlaylistURL returns the URL to access the HLS playlist
func (t *HLSTranscoder) GetPlaylistURL() string {
	// If ABR is enabled with multiple variants, return master playlist
	if t.abrConfig.Enabled && len(t.abrConfig.Variants) > 1 {
		return fmt.Sprintf("/hls/%s/master.m3u8", t.streamID)
	}

	// For single variant or non-ABR, return the variant playlist
	if t.abrConfig.Enabled && len(t.abrConfig.Variants) == 1 {
		return fmt.Sprintf("/hls/%s/%s/playlist.m3u8", t.streamID, t.abrConfig.Variants[0].Name)
	}

	// Legacy single bitrate mode
	return fmt.Sprintf("/hls/%s/playlist.m3u8", t.streamID)
}

// GetOutputDir returns the HLS output directory
func (t *HLSTranscoder) GetOutputDir() string {
	return t.hlsOutputDir
}

// IsRunning returns whether the transcoder is running
func (t *HLSTranscoder) IsRunning() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.isRunning
}

// GetFLVWriter returns the FLV writer for this transcoder
func (t *HLSTranscoder) GetFLVWriter() *FLVWriter {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.flvWriter
}

// GetStats returns transcoder statistics
func (t *HLSTranscoder) GetStats() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	stats := map[string]interface{}{
		"stream_id":     t.streamID.String(),
		"stream_key":    t.streamKey,
		"is_running":    t.isRunning,
		"bytes_written": t.bytesWritten,
		"output_dir":    t.hlsOutputDir,
		"playlist_url":  t.GetPlaylistURL(),
		"abr_enabled":   t.abrConfig.Enabled,
	}

	if t.isRunning {
		stats["duration"] = time.Since(t.startTime).Seconds()
		stats["bitrate_mbps"] = float64(t.bytesWritten*8) / time.Since(t.startTime).Seconds() / 1000000
	}

	if t.abrConfig.Enabled {
		variants := make([]string, len(t.abrConfig.Variants))
		for i, v := range t.abrConfig.Variants {
			variants[i] = v.Name
		}
		stats["variants"] = variants
	}

	return stats
}

// buildABRConfig builds ABR configuration from config
func buildABRConfig(cfg *config.Config) *ABRConfig {
	if !cfg.EnableABR {
		return &ABRConfig{
			Enabled:  false,
			Variants: nil,
		}
	}

	// Get all available variants
	allVariants := DefaultABRVariants()

	// Filter variants based on config
	var enabledVariants []ABRVariant
	if len(cfg.ABRVariants) == 0 {
		// If no specific variants configured, use all
		enabledVariants = allVariants
	} else {
		// Filter to only enabled variants
		for _, variantName := range cfg.ABRVariants {
			for _, variant := range allVariants {
				if variant.Name == variantName {
					enabledVariants = append(enabledVariants, variant)
					break
				}
			}
		}
	}

	return &ABRConfig{
		Enabled:  true,
		Variants: enabledVariants,
	}
}

// buildFilterComplex builds the FFmpeg filter_complex string for scaling video to multiple resolutions
func (t *HLSTranscoder) buildFilterComplex() string {
	if !t.abrConfig.Enabled || len(t.abrConfig.Variants) <= 1 {
		return ""
	}

	numVariants := len(t.abrConfig.Variants)

	// Split the input video stream into multiple outputs
	// Example: [0:v]split=4[v0out][v1out][v2out][v3out]
	var filterParts []string

	// Build split filter
	splitOutputs := make([]string, numVariants)
	for i := 0; i < numVariants; i++ {
		splitOutputs[i] = fmt.Sprintf("[v%dout]", i)
	}
	splitFilter := fmt.Sprintf("[0:v]split=%d%s", numVariants, strings.Join(splitOutputs, ""))
	filterParts = append(filterParts, splitFilter)

	// Build scale filters for each variant
	for i, variant := range t.abrConfig.Variants {
		scaleFilter := fmt.Sprintf("[v%dout]scale=%d:%d[v%d]", i, variant.Width, variant.Height, i)
		filterParts = append(filterParts, scaleFilter)
	}

	return strings.Join(filterParts, "; ")
}

// buildFFmpegArgs builds FFmpeg arguments for single or multi-bitrate output
func (t *HLSTranscoder) buildFFmpegArgs() []string {
	if !t.abrConfig.Enabled || len(t.abrConfig.Variants) == 0 {
		// Fall back to single bitrate mode
		return t.buildSingleBitrateArgs()
	}

	if len(t.abrConfig.Variants) == 1 {
		// Single variant - no need for complex filtering
		return t.buildSingleVariantArgs(t.abrConfig.Variants[0])
	}

	// Multi-bitrate ABR mode
	return t.buildMultiBitrateArgs()
}

// buildSingleBitrateArgs builds args for legacy single bitrate mode
func (t *HLSTranscoder) buildSingleBitrateArgs() []string {
	playlistPath := filepath.Join(t.hlsOutputDir, "playlist.m3u8")
	segmentPattern := filepath.Join(t.hlsOutputDir, "segment_%03d.ts")

	return []string{
		"-re",
		"-i", "pipe:0",
		"-c:v", "copy",
		"-c:a", "aac",
		"-b:a", "128k",
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%d", t.config.HLSSegmentTime),
		"-hls_list_size", fmt.Sprintf("%d", t.config.HLSPlaylistLength),
		"-hls_flags", "delete_segments",
		"-hls_segment_filename", segmentPattern,
		"-hls_playlist_type", "event",
		playlistPath,
	}
}

// buildSingleVariantArgs builds args for single variant (no complex filter needed)
func (t *HLSTranscoder) buildSingleVariantArgs(variant ABRVariant) []string {
	variantDir := filepath.Join(t.hlsOutputDir, variant.Name)
	playlistPath := filepath.Join(variantDir, "playlist.m3u8")
	segmentPattern := filepath.Join(variantDir, "segment_%03d.ts")

	videoCodec := "libx264"
	if t.config.HWAccelEnabled && t.config.HWAccelType != "" {
		videoCodec = t.getHWAccelCodec()
	}

	args := []string{
		"-re",
		"-i", "pipe:0",
		"-c:v", videoCodec,
		"-preset", "veryfast",
		"-b:v", variant.VideoBitrate,
		"-maxrate", CalculateMaxRate(variant.VideoBitrate),
		"-bufsize", CalculateBufSize(variant.VideoBitrate),
		"-vf", fmt.Sprintf("scale=%d:%d", variant.Width, variant.Height),
		"-c:a", "aac",
		"-b:a", variant.AudioBitrate,
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%d", t.config.HLSSegmentTime),
		"-hls_list_size", fmt.Sprintf("%d", t.config.HLSPlaylistLength),
		"-hls_flags", "delete_segments",
		"-hls_segment_filename", segmentPattern,
		"-hls_playlist_type", "event",
		playlistPath,
	}

	return args
}

// buildMultiBitrateArgs builds args for multi-bitrate ABR
func (t *HLSTranscoder) buildMultiBitrateArgs() []string {
	args := []string{
		"-re",
		"-i", "pipe:0",
	}

	// Add filter_complex for video scaling
	filterComplex := t.buildFilterComplex()
	if filterComplex != "" {
		args = append(args, "-filter_complex", filterComplex)
	}

	videoCodec := "libx264"
	if t.config.HWAccelEnabled && t.config.HWAccelType != "" {
		videoCodec = t.getHWAccelCodec()
	}

	// Add output settings for each variant
	for i, variant := range t.abrConfig.Variants {
		variantDir := filepath.Join(t.hlsOutputDir, variant.Name)
		playlistPath := filepath.Join(variantDir, "playlist.m3u8")
		segmentPattern := filepath.Join(variantDir, "segment_%03d.ts")

		// Map the scaled video stream
		args = append(args,
			"-map", fmt.Sprintf("[v%d]", i),
			"-map", "0:a",
		)

		// Video encoding settings
		args = append(args,
			"-c:v:"+fmt.Sprint(i), videoCodec,
			"-preset:v:"+fmt.Sprint(i), "veryfast",
			"-b:v:"+fmt.Sprint(i), variant.VideoBitrate,
			"-maxrate:v:"+fmt.Sprint(i), CalculateMaxRate(variant.VideoBitrate),
			"-bufsize:v:"+fmt.Sprint(i), CalculateBufSize(variant.VideoBitrate),
			"-g:v:"+fmt.Sprint(i), fmt.Sprintf("%d", t.config.HLSSegmentTime*30), // Keyframe interval (assuming 30fps)
		)

		// Audio encoding settings
		args = append(args,
			"-c:a:"+fmt.Sprint(i), "aac",
			"-b:a:"+fmt.Sprint(i), variant.AudioBitrate,
		)

		// HLS output settings
		args = append(args,
			"-f", "hls",
			"-hls_time", fmt.Sprintf("%d", t.config.HLSSegmentTime),
			"-hls_list_size", fmt.Sprintf("%d", t.config.HLSPlaylistLength),
			"-hls_flags", "delete_segments",
			"-hls_segment_filename", segmentPattern,
			"-hls_playlist_type", "event",
			playlistPath,
		)
	}

	return args
}

// getHWAccelCodec returns the appropriate hardware acceleration codec
func (t *HLSTranscoder) getHWAccelCodec() string {
	switch t.config.HWAccelType {
	case "nvenc":
		return "h264_nvenc"
	case "qsv":
		return "h264_qsv"
	case "videotoolbox":
		return "h264_videotoolbox"
	case "vaapi":
		return "h264_vaapi"
	default:
		return "libx264"
	}
}

// GenerateMasterPlaylist generates the master.m3u8 playlist for ABR
func (t *HLSTranscoder) GenerateMasterPlaylist() error {
	if !t.abrConfig.Enabled || len(t.abrConfig.Variants) <= 1 {
		return nil // No master playlist needed for single variant
	}

	masterPath := filepath.Join(t.hlsOutputDir, "master.m3u8")

	var content strings.Builder
	content.WriteString("#EXTM3U\n")
	content.WriteString("#EXT-X-VERSION:3\n")

	for _, variant := range t.abrConfig.Variants {
		content.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d,NAME=\"%s\"\n",
			variant.Bandwidth,
			variant.Width,
			variant.Height,
			variant.Name,
		))
		content.WriteString(fmt.Sprintf("%s/playlist.m3u8\n", variant.Name))
	}

	if err := os.WriteFile(masterPath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write master playlist: %w", err)
	}

	log.Printf("[Transcoder] Generated master playlist: %s", masterPath)
	return nil
}
