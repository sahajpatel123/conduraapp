//go:build darwin

package voice

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
)

// darwinSpeaker speaks text using macOS native `say` command.
type darwinSpeaker struct {
	mu    sync.Mutex
	cmd   *exec.Cmd
	voice string
	rate  int
}

// NewSpeaker creates a new platform-specific Speaker.
// On macOS it uses the native `say` command.
func NewSpeaker(voice string, rate int) Speaker {
	return &darwinSpeaker{
		voice: voice,
		rate:  rate,
	}
}

func (s *darwinSpeaker) Speak(ctx context.Context, text string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	args := []string{}
	if s.voice != "" {
		args = append(args, "-v", s.voice)
	}
	if s.rate > 0 {
		args = append(args, "-r", fmt.Sprintf("%d", s.rate))
	}
	args = append(args, text)

	s.cmd = exec.CommandContext(ctx, "say", args...) //nolint:gosec // "say" is a known safe binary
	return s.cmd.Run()
}

func (s *darwinSpeaker) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
}
