package engine_test

import (
    "context"
    "os/exec"
    "testing"
    "github.com/kelyonnnn17/flux/internal/engine"
    "github.com/stretchr/testify/assert"
)

type stubRunner struct {
    lastCmd string
    lastArgs []string
    err bool
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
        src, dst string
        wantFirst string
    }{
        {"doc.pdf", "doc.html", "pandoc"},
        {"slide.md", "slide.pdf", "pandoc"},
        {"photo.jpg", "photo.png", "imagemagick"},
        {"img.PNG", "img.webp", "imagemagick"},
        {"video.mp4", "video.mkv", "ffmpeg"},
        {"audio.mp3", "audio.wav", "ffmpeg"},
    }
    for _, tt := range tests {
        got := engine.RouteByFormat(tt.src, tt.dst)
        assert.NotEmpty(t, got, "RouteByFormat(%q, %q)", tt.src, tt.dst)
        assert.Equal(t, tt.wantFirst, got[0], "RouteByFormat(%q, %q) first engine", tt.src, tt.dst)
    }
}
