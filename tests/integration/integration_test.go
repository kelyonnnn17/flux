//go:build integration

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestIntegration_DataConversion(t *testing.T) {
	flux := mustBuildFlux(t)
	dir := t.TempDir()

	in := filepath.Join(dir, "in.json")
	out := filepath.Join(dir, "out.yaml")
	if err := os.WriteFile(in, []byte(`{"a":1,"b":"x"}`), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(flux, "convert", "-i", in, "-o", out)
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	if outb, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("flux convert failed: %v\n%s", err, outb)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("output file empty")
	}
}

func TestIntegration_FFmpeg(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
	flux := mustBuildFlux(t)
	dir := t.TempDir()

	// Create a minimal valid wav file (or use ffmpeg to create one)
	// For simplicity, create empty input and expect ffmpeg to fail gracefully
	// Better: use a real tiny media file. We'll create via ffmpeg first.
	initWav := filepath.Join(dir, "init.wav")
	initCmd := exec.Command("ffmpeg", "-f", "lavfi", "-i", "sine=frequency=440:duration=0.1", "-y", initWav)
	if err := initCmd.Run(); err != nil {
		t.Skip("could not create test wav:", err)
	}

	out := filepath.Join(dir, "out.mp3")
	cmd := exec.Command(flux, "convert", "-i", initWav, "-o", out, "--engine", "ffmpeg")
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	if outb, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("flux convert ffmpeg failed: %v\n%s", err, outb)
	}

	if _, err := os.Stat(out); err != nil {
		t.Fatal("output file not created:", err)
	}
}

func TestIntegration_ImageMagick(t *testing.T) {
	if _, err := exec.LookPath("magick"); err != nil {
		if _, err := exec.LookPath("convert"); err != nil {
			t.Skip("imagemagick not installed")
		}
	}
	flux := mustBuildFlux(t)
	dir := t.TempDir()

	inPath := filepath.Join(dir, "in.png")
	bin := "magick"
	if _, err := exec.LookPath("magick"); err != nil {
		bin = "convert"
	}
	createCmd := exec.Command(bin, "xc:red", inPath)
	if err := createCmd.Run(); err != nil {
		t.Skip("could not create test image:", err)
	}

	out := filepath.Join(dir, "out.jpg")
	cmd := exec.Command(flux, "convert", "-i", inPath, "-o", out, "--engine", "imagemagick")
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	if outb, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("flux convert imagemagick failed: %v\n%s", err, outb)
	}

	if _, err := os.Stat(out); err != nil {
		t.Fatal("output file not created:", err)
	}
}

func TestIntegration_Pandoc(t *testing.T) {
	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc not installed")
	}
	flux := mustBuildFlux(t)
	dir := t.TempDir()

	in := filepath.Join(dir, "in.md")
	out := filepath.Join(dir, "out.html")
	if err := os.WriteFile(in, []byte("# Hello\n\nWorld"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(flux, "convert", "-i", in, "-o", out, "--engine", "pandoc")
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	if outb, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("flux convert pandoc failed: %v\n%s", err, outb)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("output file empty")
	}
}

func mustBuildFlux(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	out := filepath.Join(dir, "flux")
	cmd := exec.Command("go", "build", "-o", out, ".")
	cmd.Dir = filepath.Join("..", "..")
	if err := cmd.Run(); err != nil {
		t.Fatal("build flux:", err)
	}
	return out
}
