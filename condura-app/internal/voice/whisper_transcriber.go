package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

const (
	whisperSampleRate = 16000
	whisperChannels   = 1
)

// whisperTranscriber wraps the whisper-cli binary for transcription.
type whisperTranscriber struct {
	binaryPath string
	modelPath  string
	language   string
	mu         sync.Mutex
}

// NewTranscriber creates a new whisper-based Transcriber.
func NewTranscriber(binaryPath, modelPath, language string) Transcriber {
	return &whisperTranscriber{
		binaryPath: binaryPath,
		modelPath:  modelPath,
		language:   language,
	}
}

// whisperOutput is the JSON structure from whisper-cli --output-json.
type whisperOutput struct {
	Result struct {
		Text string `json:"text"`
	} `json:"result"`
	Segments []struct {
		Start float64 `json:"start"`
		End   float64 `json:"end"`
		Text  string  `json:"text"`
	} `json:"segments"`
}

func (t *whisperTranscriber) Transcribe(ctx context.Context, audio []byte) (Transcript, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Write audio to a temporary file.
	tmpFile, err := createTempWAV(audio)
	if err != nil {
		return Transcript{}, fmt.Errorf("create temp file: %w", err)
	}
	defer func() { cleanupTempFile(tmpFile) }()

	args := []string{
		"-m", t.modelPath,
		"-f", tmpFile,
		"--output-json",
	}

	if t.language != "" && t.language != "auto" {
		args = append(args, "-l", t.language)
	}

	cmd := exec.CommandContext(ctx, t.binaryPath, args...) //nolint:gosec // binaryPath is from config
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return Transcript{}, fmt.Errorf("whisper-cli: %w\n%s", err, stderr.String())
	}

	var output whisperOutput
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		return Transcript{}, fmt.Errorf("parse whisper output: %w", err)
	}

	segments := make([]Segment, len(output.Segments))
	for i, s := range output.Segments {
		segments[i] = Segment{
			Start: s.Start,
			End:   s.End,
			Text:  s.Text,
		}
	}

	return Transcript{
		Text:     output.Result.Text,
		Language: t.language,
		Segments: segments,
	}, nil
}

func (t *whisperTranscriber) TranscribeStream(ctx context.Context, audio <-chan []float32) (<-chan Partial, error) {
	out := make(chan Partial, 100)

	go func() {
		defer close(out)

		// Collect all samples until channel closes.
		var allSamples []float32
		for samples := range audio {
			allSamples = append(allSamples, samples...)
		}

		if len(allSamples) == 0 {
			return
		}

		// Convert to WAV and transcribe.
		wav := encodeWAV(allSamples, whisperSampleRate, whisperChannels)
		transcript, err := t.Transcribe(ctx, wav)
		if err != nil {
			return
		}

		// Emit segments as partials, last one is final.
		for i, seg := range transcript.Segments {
			out <- Partial{
				Text:    seg.Text,
				IsFinal: i == len(transcript.Segments)-1,
			}
		}
	}()

	return out, nil
}

func createTempWAV(data []byte) (string, error) {
	tmpFile, err := createTempFile()
	if err != nil {
		return "", err
	}

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return "", err
	}

	return tmpFile.Name(), tmpFile.Close()
}

func createTempFile() (*os.File, error) {
	return os.CreateTemp("", "whisper-input-*.wav")
}

func cleanupTempFile(path string) {
	_ = os.Remove(path)
}

// FindWhisperBinary locates the whisper-cli binary.
// Returns the path if found, empty string otherwise.
func FindWhisperBinary() string {
	// Check common locations.
	locations := []string{
		"/usr/local/bin/whisper-cli",
		"/opt/homebrew/bin/whisper-cli",
		filepath.Join(userHomeDir(), ".local/bin/whisper-cli"),
	}

	for _, loc := range locations {
		if _, err := exec.LookPath(loc); err == nil {
			return loc
		}
	}

	// Try PATH.
	if path, err := exec.LookPath("whisper-cli"); err == nil {
		return path
	}

	return ""
}

func userHomeDir() string {
	if home, err := os.UserHomeDir(); err == nil {
		return home
	}
	return ""
}
