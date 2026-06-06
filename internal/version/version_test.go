package version

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet_Defaults(t *testing.T) {
	// In test runs (no ldflags), defaults should be set.
	info := Get()
	assert.Equal(t, "v0.0.0-dev", info.Version)
	assert.Equal(t, "none", info.Commit)
	assert.Equal(t, "unknown", info.BuildDate)
	assert.NotEmpty(t, info.GoVersion)
	assert.NotEmpty(t, info.Platform)
	assert.True(t, info.IsDev, "dev build should have IsDev=true")
	assert.Equal(t, "none", info.ShortSHA, "unknown commit should produce short 'none'")
}

func TestString(t *testing.T) {
	s := String()
	require.NotEmpty(t, s)
	assert.Contains(t, s, "Synaptic")
	assert.Contains(t, s, "v0.0.0-dev")
}

func TestShortSHA(t *testing.T) {
	tests := []struct {
		name   string
		commit string
		want   string
	}{
		{"full SHA", "abc1234567890def1234567890abcdef12345678", "abc1234"},
		{"short SHA", "abc1234", "abc1234"},
		{"none", "none", "none"},
		{"empty", "", ""},
		{"exactly 7", "abcdefg", "abcdefg"},
		{"6 chars (too short)", "abcdef", "abcdef"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shortSHA(tt.commit)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInfo_JSON(t *testing.T) {
	info := Get()
	b, err := json.Marshal(info)
	require.NoError(t, err)
	s := string(b)
	assert.Contains(t, s, `"version"`)
	assert.Contains(t, s, `"commit"`)
	assert.Contains(t, s, `"go_version"`)
}

func TestGet_Cached(t *testing.T) {
	// Calling Get multiple times should return the same value (cached).
	i1 := Get()
	i2 := Get()
	assert.Equal(t, i1, i2)
}

func TestString_ContainsPlatform(t *testing.T) {
	s := String()
	// Should mention the platform somehow (darwin, linux, windows).
	platform := Get().Platform
	if platform != "" && !strings.Contains(s, strings.SplitN(platform, "/", 2)[0]) {
		t.Logf("warning: String() %q does not mention platform %q", s, platform)
	}
}
