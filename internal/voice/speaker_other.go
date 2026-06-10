//go:build !darwin

package voice

import "context"

// noopSpeaker is a Speaker stub for non-darwin platforms.
type noopSpeaker struct{}

// NewSpeaker creates a platform-specific Speaker. On non-darwin
// platforms this is a noop; Speak returns an error.
func NewSpeaker(_ string, _ int) Speaker {
	return &noopSpeaker{}
}

func (s *noopSpeaker) Speak(_ context.Context, _ string) error {
	return fmtErrNotImplemented
}

func (s *noopSpeaker) Stop() {}

// fmtErrNotImplemented is returned by the noop speaker.
var fmtErrNotImplemented = errNotImplemented{}

type errNotImplemented struct{}

func (errNotImplemented) Error() string {
	return "TTS not yet implemented on this platform"
}
