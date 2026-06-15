package replay

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestExportMP4_NoFrames(t *testing.T) {
	_, err := ExportMP4(context.Background(), nil, t.TempDir()+"/out.mp4")
	if err == nil {
		t.Fatal("expected error for empty frames")
	}
}

func TestExportMP4_NoScreenshots(t *testing.T) {
	frames := []Frame{{Outcome: OutcomeUnknown}}
	_, err := ExportMP4(context.Background(), frames, t.TempDir()+"/out.mp4")
	if err == nil {
		t.Fatal("expected error when no screenshots")
	}
}

func TestWriteFramePNG(t *testing.T) {
	dir := t.TempDir()
	pngBytes := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	if err := writeFramePNG(dir, 1, pngBytes); err != nil {
		t.Fatal(err)
	}
}

func TestExportMP4_Integration(t *testing.T) {
	if _, err := execLookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
	dir := t.TempDir()
	pngBytes := encodeTestPNG(t, 4, 4)
	frames := []Frame{{BeforeScreenshot: pngBytes}}
	out := filepath.Join(dir, "test.mp4")
	path, err := ExportMP4(context.Background(), frames, out)
	if err != nil {
		t.Fatalf("ExportMP4: %v", err)
	}
	if path != out {
		t.Fatalf("path %s", path)
	}
	st, err := os.Stat(out)
	if err != nil || st.Size() < 100 {
		t.Fatalf("mp4 missing or tiny: %v", err)
	}
}

var execLookPath = exec.LookPath

func encodeTestPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 40), uint8(y * 40), 128, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
