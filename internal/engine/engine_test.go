package engine_test

import (
	"context"
	"github.com/kelyonnnn17/flux/internal/engine"
	"github.com/stretchr/testify/assert"
	"os/exec"
	"testing"
)

type stubRunner struct {
	lastCmd  string
	lastArgs []string
}

func (s *stubRunner) CommandContext(ctx context.Context, name string, arg ...string) *exec.Cmd {
	s.lastCmd = name
	s.lastArgs = arg
	// Return a harmless command that succeeds
	return exec.CommandContext(ctx, "true")
}

func TestFFmpegAdapter(t *testing.T) {
	r := &stubRunner{}
	a := &engine.FFmpegAdapter{Runner: r}
	err := a.Convert("in.mp4", "out.mp3", []string{"-b:a", "192k"})
	assert.NoError(t, err)
	assert.Equal(t, "ffmpeg", r.lastCmd)
}

func TestImageMagickAdapter(t *testing.T) {
	r := &stubRunner{}
	a := &engine.ImageMagickAdapter{Runner: r}
	err := a.Convert("in.png", "out.jpg", []string{"-resize", "100x100"})
	assert.NoError(t, err)
	// binary may be magick or convert; accept either
	assert.Contains(t, []string{"magick", "convert"}, r.lastCmd)
}

func TestPandocAdapter(t *testing.T) {
	r := &stubRunner{}
	a := &engine.PandocAdapter{Runner: r}
	err := a.Convert("in.md", "out.pdf", nil)
	assert.NoError(t, err)
	assert.Equal(t, "pandoc", r.lastCmd)
}

func TestRouteByFormat(t *testing.T) {
	tests := []struct {
		src, dst  string
		wantFirst string
	}{
		{"doc.pdf", "doc.html", "pandoc"},
		{"slide.md", "slide.pdf", "pandoc"},
		{"photo.jpg", "photo.png", "imagemagick"},
		{"img.PNG", "img.webp", "imagemagick"},
		{"video.mp4", "video.mkv", "ffmpeg"},
		{"audio.mp3", "audio.wav", "ffmpeg"},
		{"data.json", "data.yaml", "data"},
		{"in.csv", "out.json", "data"},
	}
	for _, tt := range tests {
		got := engine.RouteByFormat(tt.src, tt.dst)
		assert.NotEmpty(t, got, "RouteByFormat(%q, %q)", tt.src, tt.dst)
		assert.Equal(t, tt.wantFirst, got[0], "RouteByFormat(%q, %q) first engine", tt.src, tt.dst)
	}
}

func TestCanConvert_SupportedFormats(t *testing.T) {
	tests := []struct {
		src, dst string
		wantErr  bool
	}{
		// Pandoc can read/write documents (except PDF input)
		{"input.md", "output.docx", false},
		{"input.docx", "output.pdf", false},
		{"input.html", "output.md", false},
		// ImageMagick can read/write images
		{"input.jpg", "output.png", false},
		{"input.png", "output.gif", false},
		// FFmpeg can read/write audio/video
		{"input.mp3", "output.wav", false},
		{"input.mp4", "output.mkv", false},
		// Data formats
		{"input.json", "output.yaml", false},
		{"input.csv", "output.json", false},
	}
	for _, tt := range tests {
		_, _, err := engine.CanConvert(tt.src, tt.dst)
		if tt.wantErr {
			assert.Error(t, err, "CanConvert(%q, %q) expected error", tt.src, tt.dst)
		} else {
			assert.NoError(t, err, "CanConvert(%q, %q) unexpected error", tt.src, tt.dst)
		}
	}
}

func TestCanConvert_UnsupportedFormats(t *testing.T) {
	tests := []struct {
		src, dst    string
		wantErr     bool
		wantSuggest bool // Should include a workaround suggestion
	}{
		// PDF input is not supported by Pandoc
		{"input.pdf", "output.docx", true, true},
		// Same source and dest format
		{"input.jpg", "output.jpg", true, false},
		// Impossible conversions
		{"input.json", "output.mp4", true, false},
	}
	for _, tt := range tests {
		_, suggest, err := engine.CanConvert(tt.src, tt.dst)
		if tt.wantErr {
			assert.Error(t, err, "CanConvert(%q, %q) expected error", tt.src, tt.dst)
		} else {
			assert.NoError(t, err, "CanConvert(%q, %q) unexpected error", tt.src, tt.dst)
		}
		if tt.wantSuggest {
			assert.NotEmpty(t, suggest, "CanConvert(%q, %q) expected suggestion", tt.src, tt.dst)
		}
	}
}
