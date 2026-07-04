//go:build darwin && cgo

package voice

import (
	"context"
	"fmt"
	"sync"
	"unsafe"

	"github.com/gen2brain/malgo"
)

// darwinRecorder captures audio from the microphone using malgo
// (CoreAudio on macOS). The device is initialized in Start and
// torn down in Stop. PCM float32 samples are pushed into the
// samplesCh channel for live waveform/VAD consumption, and
// accumulated in captured for the final WAV encoding.
type darwinRecorder struct {
	sampleRate uint32
	channels   uint32
	samplesCh  chan []float32

	mu       sync.Mutex
	captured []float32
	device   *malgo.Device
	maCtx    *malgo.AllocatedContext
}

// NewRecorder creates a platform-specific Recorder backed by malgo.
func NewRecorder(sampleRate, channels int) Recorder {
	return &darwinRecorder{
		sampleRate: uint32(sampleRate), //nolint:gosec // sampleRate is validated by config
		channels:   uint32(channels),   //nolint:gosec // channels is validated by config
		samplesCh:  make(chan []float32, 100),
	}
}

// RecorderAvailable reports whether the platform can capture audio.
// On macOS this checks that malgo can initialize the CoreAudio backend
// and enumerate at least one capture device.
func RecorderAvailable() bool {
	maCtx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return false
	}
	defer maCtx.Free()
	devs, err := maCtx.Devices(malgo.Capture)
	if err != nil || len(devs) == 0 {
		return false
	}
	return true
}

func (r *darwinRecorder) Start(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.captured = nil

	// Initialize the malgo context with default backends.
	maCtx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(msg string) {
		// malgo log callback; intentionally quiet in production.
		_ = msg
	})
	if err != nil {
		return fmt.Errorf("voice: malgo init: %w", err)
	}
	r.maCtx = maCtx

	// Configure the capture device.
	cfg := malgo.DefaultDeviceConfig(malgo.Capture)
	cfg.SampleRate = r.sampleRate
	cfg.Capture.Format = malgo.FormatF32
	cfg.Capture.Channels = r.channels
	cfg.PerformanceProfile = malgo.LowLatency
	cfg.NoPreSilencedOutputBuffer = 1
	cfg.NoFixedSizedCallback = 1

	// The data callback receives raw PCM bytes from the audio device.
	// For FormatF32, every 4 bytes are one float32 sample.
	callbackFn := malgo.DataProc(func(_, inputSamples []byte, framecount uint32) {
		if len(inputSamples) == 0 {
			return
		}
		// Convert bytes to float32 samples.
		n := len(inputSamples) / 4 // 4 bytes per float32
		samples := make([]float32, n)
		for i := range samples {
			// Reinterpret 4 bytes as float32 (little-endian on macOS).
			samples[i] = float32FromBytes(inputSamples[i*4 : i*4+4])
		}

		r.mu.Lock()
		r.captured = append(r.captured, samples...)
		r.mu.Unlock()

		// Push a copy to the samples channel for live waveform/VAD.
		// Non-blocking: if the consumer is slow, drop this chunk.
		cp := make([]float32, len(samples))
		copy(cp, samples)
		select {
		case r.samplesCh <- cp:
		default:
		}
	})

	device, err := malgo.InitDevice(maCtx.Context, cfg, malgo.DeviceCallbacks{
		Data: callbackFn,
	})
	if err != nil {
		maCtx.Free()
		r.maCtx = nil
		return fmt.Errorf("voice: malgo device init: %w", err)
	}
	r.device = device

	if err := device.Start(); err != nil {
		device.Uninit()
		maCtx.Free()
		r.device = nil
		r.maCtx = nil
		return fmt.Errorf("voice: malgo device start: %w", err)
	}

	// Block until ctx is canceled. The caller (pipeline) manages the
	// lifecycle: Start blocks, then Stop is called from another goroutine.
	<-ctx.Done()
	return nil
}

func (r *darwinRecorder) Stop() ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Tear down the malgo device and context.
	if r.device != nil {
		_ = r.device.Stop()
		r.device.Uninit()
		r.device = nil
	}
	if r.maCtx != nil {
		r.maCtx.Free()
		r.maCtx = nil
	}

	samples := r.captured
	r.captured = nil
	return encodeWAV(samples, r.sampleRate, r.channels), nil
}

func (r *darwinRecorder) Samples() <-chan []float32 {
	return r.samplesCh
}

// float32FromBytes reinterprets 4 bytes as a float32 (little-endian).
func float32FromBytes(b []byte) float32 {
	// Use unsafe to reinterpret the bytes as float32. This is safe
	// because we know the input is exactly 4 bytes and the platform
	// is little-endian (macOS ARM64/x86_64).
	return *(*float32)(unsafe.Pointer(&b[0])) //nolint:gosec // safe: 4-byte alignment guaranteed by malgo
}
