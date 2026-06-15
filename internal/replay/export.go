package replay

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ExportMP4 writes an H.264 MP4 slideshow from timeline frames using ffmpeg.
// Each frame contributes its before screenshot (if any) then after screenshot.
// Returns the output path on success.
func ExportMP4(ctx context.Context, frames []Frame, dest string) (string, error) {
	if len(frames) == 0 {
		return "", fmt.Errorf("replay: no frames to export")
	}
	if dest == "" {
		home, _ := os.UserHomeDir()
		dest = filepath.Join(home, "Documents", fmt.Sprintf("synaptic-replay-%s.mp4", time.Now().UTC().Format("20060102-150405")))
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
		return "", err
	}
	if !strings.HasSuffix(strings.ToLower(dest), ".mp4") {
		dest += ".mp4"
	}

	tmpDir, err := os.MkdirTemp("", "synaptic-replay-export-*")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	idx := 1
	for _, f := range frames {
		if len(f.BeforeScreenshot) > 0 {
			if err := writeFramePNG(tmpDir, idx, f.BeforeScreenshot); err != nil {
				return "", err
			}
			idx++
		}
		if len(f.AfterScreenshot) > 0 {
			if err := writeFramePNG(tmpDir, idx, f.AfterScreenshot); err != nil {
				return "", err
			}
			idx++
		}
	}
	if idx == 1 {
		return "", fmt.Errorf("replay: no screenshots in timeline window")
	}

	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", fmt.Errorf("replay: ffmpeg not found in PATH (install ffmpeg to export MP4)")
	}

	pattern := filepath.Join(tmpDir, "%04d.png")
	args := []string{
		"-y",
		"-framerate", "2",
		"-i", pattern,
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		dest,
	}
	cmd := exec.CommandContext(ctx, ffmpeg, args...) //nolint:gosec // ffmpeg path from LookPath; args are fixed flags
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("replay: ffmpeg: %w", err)
	}
	return dest, nil
}

func writeFramePNG(dir string, index int, png []byte) error {
	path := filepath.Join(dir, fmt.Sprintf("%04d.png", index))
	return os.WriteFile(path, png, 0o600)
}

// ExportMP4FromTimeline loads frames from the replay timeline and exports MP4.
func (r *Replay) ExportMP4FromTimeline(ctx context.Context, since time.Time, dest string) (string, error) {
	frames, err := r.Timeline(ctx, since)
	if err != nil {
		return "", err
	}
	return ExportMP4(ctx, frames, dest)
}
