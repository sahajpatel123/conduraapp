// Package voice provides speech-to-text and text-to-speech for Synaptic.
//
// The package defines three core interfaces:
//   - Recorder: captures audio from the microphone
//   - Transcriber: converts audio bytes to text (via whisper subprocess)
//   - Speaker: converts text to speech using OS-native TTS
//
// Concrete implementations are platform-specific (build-tagged) and
// depend on external tools (whisper-cli binary, OS TTS engines).
package voice
