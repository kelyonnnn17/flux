package engine_test

import (
	"context"
	"os/exec"
	"testing"

	"github.com/kelyonnnn17/flux/internal/engine"
	"github.com/stretchr/testify/assert"
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

func TestPythonPDFAdapter_UnknownMode(t *testing.T) {
	a := &engine.PythonPDFAdapter{Runner: &stubRunner{}, Mode: "unknown"}
	err := a.Convert("in", "out", nil)
	assert.Error(t, err)
}

func TestRouteByFormat(t *testing.T) {
	tests := []struct {
		src, dst  string
		wantFirst string
	}{
		{"doc.pdf", "doc.html", "pandoc"},
		{"doc.pdf", "doc.docx", "pdf2docx"},
		{"doc.docx", "doc.pdf", "docx2pdf"},
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
		// Same source and dest format
		{"input.jpg", "output.jpg", true, false},
		// Same source and dest format
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

func TestPlanConversion_DirectPreferred(t *testing.T) {
	route, err := engine.PlanConversion("input.docx", "output.pdf", "auto")
	assert.NoError(t, err)
	assert.Len(t, route.Steps, 1)
	assert.Contains(t, []string{"docx2pdf", "pandoc"}, route.Steps[0].Engine)
}

func TestPlanConversion_PDFBestEffortRoute(t *testing.T) {
	route, err := engine.PlanConversion("input.pdf", "output.docx", "auto")
	assert.NoError(t, err)
	assert.NotEmpty(t, route.Steps)
	if route.Steps[0].Engine == "pdf2docx" {
		assert.Len(t, route.Steps, 1)
		assert.Equal(t, "pdf", route.Steps[0].FromFormat)
		assert.Equal(t, "docx", route.Steps[0].ToFormat)
		assert.Empty(t, route.Warnings)
		return
	}

	assert.GreaterOrEqual(t, len(route.Steps), 2)
	assert.Equal(t, "pdftotext", route.Steps[0].Engine)
	assert.Equal(t, "pdf", route.Steps[0].FromFormat)
	assert.Equal(t, "txt", route.Steps[0].ToFormat)
	assert.Equal(t, "pandoc", route.Steps[len(route.Steps)-1].Engine)
	assert.NotEmpty(t, route.Warnings)
}

func TestPlanConversion_PDFBestEffortRoute_ForcedPandoc(t *testing.T) {
	route, err := engine.PlanConversion("input.pdf", "output.docx", "pandoc")
	assert.Error(t, err)
	assert.Empty(t, route.Steps)
}

func TestCanEngineConvert_MultiHopForcedEngine(t *testing.T) {
	ok, err := engine.CanEngineConvert("input.rst", "output.docx", "pandoc")
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = engine.CanEngineConvert("input.pdf", "output.docx", "pandoc")
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestCanEngineConvert(t *testing.T) {
	tests := []struct {
		src, dst   string
		engineName string
		wantOk     bool
		wantErr    bool
	}{
		{"input.jpg", "output.png", "pandoc", false, false},
		{"input.jpg", "output.png", "imagemagick", true, false},
		{"input.json", "output.yaml", "data", true, false},
		{"input.json", "output.yaml", "ffmpeg", false, false},
		{"input.jpg", "output.jpg", "ffmpeg", false, true}, // same extension
		{"input.jpg", "output.png", "unknown", false, true},
	}

	for _, tt := range tests {
		ok, err := engine.CanEngineConvert(tt.src, tt.dst, tt.engineName)
		if tt.wantErr {
			assert.Error(t, err, "CanEngineConvert(%q, %q, %q) expected error", tt.src, tt.dst, tt.engineName)
		} else {
			assert.NoError(t, err, "CanEngineConvert(%q, %q, %q) unexpected error", tt.src, tt.dst, tt.engineName)
		}
		assert.Equal(t, tt.wantOk, ok, "CanEngineConvert(%q, %q, %q) ok", tt.src, tt.dst, tt.engineName)
	}
}
