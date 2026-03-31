package thumbnail

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) GenerateFromVideo(videoReader io.Reader, videoSize int64) (io.Reader, int64, error) {
	tmpDir, err := os.MkdirTemp("", "video-thumbnail-*")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	videoPath := filepath.Join(tmpDir, "video.tmp")
	thumbnailPath := filepath.Join(tmpDir, "thumbnail.jpg")

	videoFile, err := os.Create(videoPath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create temp video file: %w", err)
	}

	if _, err := io.Copy(videoFile, videoReader); err != nil {
		videoFile.Close()
		return nil, 0, fmt.Errorf("failed to write video to temp file: %w", err)
	}
	videoFile.Close()

	cmd := exec.Command(
		"ffmpeg",
		"-i", videoPath,
		"-vframes", "1",
		"-vf", "scale=320:-1",
		"-f", "image2",
		"-y",
		thumbnailPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, 0, fmt.Errorf("ffmpeg failed: %w, stderr: %s", err, stderr.String())
	}

	thumbnailData, err := os.ReadFile(thumbnailPath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read thumbnail: %w", err)
	}

	return bytes.NewReader(thumbnailData), int64(len(thumbnailData)), nil
}
