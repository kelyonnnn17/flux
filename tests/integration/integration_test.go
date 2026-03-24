//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// ============================================================================
// DOCUMENT FORMATTING TESTS (Pandoc)
// ============================================================================

func TestIntegration_Pandoc_MDtoHTML_Formatted(t *testing.T) {
	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc not installed")
	}
	flux := mustBuildFlux(t)
	dir := t.TempDir()

	markdown := `# Main Title
## Subtitle

Regular paragraph with **bold** and *italic*.

### Code Block
` + "`" + `python
def hello():
    print("world")
` + "`" + `

- Item 1
- Item 2
- Item 3

1. First
2. Second

| Header 1 | Header 2 |
|----------|----------|
| Value 1  | Value 2  |
`

	in := filepath.Join(dir, "in.md")
	out := filepath.Join(dir, "out.html")
	if err := os.WriteFile(in, []byte(markdown), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(flux, "convert", "-i", in, "-o", out, "--engine", "pandoc")
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	if outb, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("MD→HTML conversion failed: %v\n%s", err, outb)
	}

	html, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}

	htmlStr := string(html)

	// Validate structure preservation
	checks := map[string]string{
		"h1 tag":       "<h1",
		"h2 tag":       "<h2",
		"h3 tag":       "<h3",
		"paragraph":    "<p>",
		"bold":         "<strong>",
		"italic":       "<em>",
		"code block":   "<code>",
		"list":         "<li>",
		"table":        "<table>",
	}

	for checkName, pattern := range checks {
		if !strings.Contains(htmlStr, pattern) {
			t.Errorf("MD→HTML: missing %s (pattern: %s)", checkName, pattern)
		}
	}

	// Validate formatter effects: standalone document should be present.
	if !strings.Contains(htmlStr, "<html") {
		t.Error("MD→HTML: missing standalone HTML wrapper (<html>)")
	}
	if !strings.Contains(htmlStr, "<body") {
		t.Error("MD→HTML: missing standalone HTML body (<body>)")
	}

	// Validate content is present
	if !strings.Contains(htmlStr, "Main Title") {
		t.Error("MD→HTML: content 'Main Title' missing")
	}
	if !strings.Contains(htmlStr, "hello") {
		t.Error("MD→HTML: code content missing")
	}
}

func TestIntegration_Pandoc_MDtoPDF_Formatted(t *testing.T) {
	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc not installed")
	}
	// Also check for PDF engine (pdflatex or xelatex)
	if _, err := exec.LookPath("pdflatex"); err != nil {
		if _, err := exec.LookPath("xelatex"); err != nil {
			t.Skip("pdflatex/xelatex not installed")
		}
	}

	flux := mustBuildFlux(t)
	dir := t.TempDir()

	markdown := "# Document\n\nThis is a test document with content that should be in the PDF.\n\n- Point 1\n- Point 2"

	in := filepath.Join(dir, "in.md")
	out := filepath.Join(dir, "out.pdf")
	if err := os.WriteFile(in, []byte(markdown), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(flux, "convert", "-i", in, "-o", out, "--engine", "pandoc")
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	if outb, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("MD→PDF conversion failed: %v\n%s", err, outb)
	}

	stat, err := os.Stat(out)
	if err != nil {
		t.Fatal("PDF file not created:", err)
	}

	// Validate PDF magic bytes
	pdfBytes := make([]byte, 4)
	f, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.Read(pdfBytes); err != nil {
		t.Fatal(err)
	}
	if string(pdfBytes[:4]) != "%PDF" {
		t.Error("PDF magic bytes not found; file may be corrupted")
	}

	// PDF should be non-trivial size (has content)
	if stat.Size() < 1000 {
		t.Errorf("PDF too small (%d bytes), likely missing content", stat.Size())
	}
}

func TestIntegration_Pandoc_HTMLtoMD(t *testing.T) {
	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc not installed")
	}
	flux := mustBuildFlux(t)
	dir := t.TempDir()

	html := `<h1>Test Heading</h1>
<p>This is a paragraph.</p>
<ul>
<li>Item 1</li>
<li>Item 2</li>
</ul>
<table>
<tr><th>Col1</th><th>Col2</th></tr>
<tr><td>A</td><td>B</td></tr>
</table>`

	in := filepath.Join(dir, "in.html")
	out := filepath.Join(dir, "out.md")
	if err := os.WriteFile(in, []byte(html), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(flux, "convert", "-i", in, "-o", out, "--engine", "pandoc")
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	if outb, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("HTML→MD conversion failed: %v\n%s", err, outb)
	}

	md, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}

	mdStr := string(md)

	// Validate markdown structure is present
	if !strings.Contains(mdStr, "# ") && !strings.Contains(mdStr, "Test Heading") {
		t.Error("HTML→MD: heading not converted")
	}
	if !strings.Contains(mdStr, "-") && !strings.Contains(mdStr, "Item 1") {
		t.Error("HTML→MD: list items missing")
	}
}

func TestIntegration_Pandoc_MDtoDOCX_Formatted(t *testing.T) {
	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc not installed")
	}
	flux := mustBuildFlux(t)
	dir := t.TempDir()

	markdown := `# Heading 1
## Heading 2

Paragraph with **bold** and *italic*.

- Bullet 1
- Bullet 2

1. Numbered 1
2. Numbered 2`

	in := filepath.Join(dir, "in.md")
	out := filepath.Join(dir, "out.docx")
	if err := os.WriteFile(in, []byte(markdown), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(flux, "convert", "-i", in, "-o", out, "--engine", "pandoc")
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	if outb, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("MD→DOCX conversion failed: %v\n%s", err, outb)
	}

	stat, err := os.Stat(out)
	if err != nil {
		t.Fatal("DOCX file not created:", err)
	}

	// DOCX should be reasonable size (at least 1KB, ZIP format has overhead)
	if stat.Size() < 1000 {
		t.Errorf("DOCX file too small (%d bytes)", stat.Size())
	}
}

// ============================================================================
// DATA FORMAT PRESERVATION TESTS
// ============================================================================

func TestIntegration_Data_JSONtoYAML_Formatted(t *testing.T) {
	flux := mustBuildFlux(t)
	dir := t.TempDir()

	jsonData := `{"name":"John","age":30,"items":["a","b"],"nested":{"key":"value"}}`

	in := filepath.Join(dir, "in.json")
	out := filepath.Join(dir, "out.yaml")
	if err := os.WriteFile(in, []byte(jsonData), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(flux, "convert", "-i", in, "-o", out)
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	if outb, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("JSON→YAML conversion failed: %v\n%s", err, outb)
	}

	yaml, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}

	yamlStr := string(yaml)

	// Validate all key values are present
	checks := []string{"John", "30", "items", "nested", "value"}
	for _, check := range checks {
		if !strings.Contains(yamlStr, check) {
			t.Errorf("JSON→YAML: value '%s' missing", check)
		}
	}
}

func TestIntegration_Data_Roundtrip(t *testing.T) {
	flux := mustBuildFlux(t)
	dir := t.TempDir()

	originalJSON := `{"users":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}],"count":2}`

	// JSON → YAML
	in := filepath.Join(dir, "in.json")
	yaml := filepath.Join(dir, "mid.yaml")
	outJSON := filepath.Join(dir, "out.json")

	if err := os.WriteFile(in, []byte(originalJSON), 0644); err != nil {
		t.Fatal(err)
	}

	cmd1 := exec.Command(flux, "convert", "-i", in, "-o", yaml)
	cmd1.Env = append(os.Environ(), "NO_COLOR=1")
	if outb, err := cmd1.CombinedOutput(); err != nil {
		t.Fatalf("JSON→YAML failed: %v\n%s", err, outb)
	}

	// YAML → JSON
	cmd2 := exec.Command(flux, "convert", "-i", yaml, "-o", outJSON)
	cmd2.Env = append(os.Environ(), "NO_COLOR=1")
	if outb, err := cmd2.CombinedOutput(); err != nil {
		t.Fatalf("YAML→JSON failed: %v\n%s", err, outb)
	}

	// Validate structure is preserved (parse both as JSON)
	var original, result map[string]interface{}
	if err := json.Unmarshal([]byte(originalJSON), &original); err != nil {
		t.Fatal(err)
	}

	finalJSON, err := os.ReadFile(outJSON)
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(finalJSON, &result); err != nil {
		t.Fatalf("Result JSON invalid: %v", err)
	}

	// Check key values match
	if original["count"] != result["count"] {
		t.Errorf("Roundtrip failed: count mismatch %v != %v", original["count"], result["count"])
	}
}

// ============================================================================
// IMAGE FORMAT CONVERSION TESTS
// ============================================================================

func TestIntegration_Image_MultiFormat(t *testing.T) {
	if _, err := exec.LookPath("magick"); err != nil {
		if _, err := exec.LookPath("convert"); err != nil {
			t.Skip("imagemagick not installed")
		}
	}
	flux := mustBuildFlux(t)
	dir := t.TempDir()

	// Create test image (blue)
	in := filepath.Join(dir, "in.png")
	bin := "magick"
	if _, err := exec.LookPath("magick"); err != nil {
		bin = "convert"
	}

	createCmd := exec.Command(bin, "xc:blue", "-resize", "100x100", in)
	if err := createCmd.Run(); err != nil {
		t.Skip("could not create test image:", err)
	}

	// Test conversions: PNG → JPG → WebP
	tests := []struct {
		src string
		dst string
		fmt string
	}{
		{in, filepath.Join(dir, "out.jpg"), "JPG"},
		{filepath.Join(dir, "out.jpg"), filepath.Join(dir, "out.webp"), "WebP"},
		{filepath.Join(dir, "out.webp"), filepath.Join(dir, "final.png"), "PNG"},
	}

	for _, tc := range tests {
		if _, err := os.Stat(tc.src); err != nil && tc.src != in {
			t.Fatalf("Source file missing for %s conversion", tc.fmt)
		}

		cmd := exec.Command(flux, "convert", "-i", tc.src, "-o", tc.dst, "--engine", "imagemagick")
		cmd.Env = append(os.Environ(), "NO_COLOR=1")
		if outb, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s conversion failed: %v\n%s", tc.fmt, err, outb)
		}

		stat, err := os.Stat(tc.dst)
		if err != nil {
			t.Fatalf("%s output file not created", tc.fmt)
		}
		if stat.Size() == 0 {
			t.Errorf("%s file is empty", tc.fmt)
		}
	}
}

// ============================================================================
// AUDIO/VIDEO CODEC CONVERSION TESTS
// ============================================================================

func TestIntegration_Audio_MultiCodec(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
	flux := mustBuildFlux(t)
	dir := t.TempDir()

	// Create test audio (1 second WAV)
	wav := filepath.Join(dir, "test.wav")
	createCmd := exec.Command("ffmpeg", "-f", "lavfi", "-i", "sine=frequency=440:duration=1", "-y", wav)
	if err := createCmd.Run(); err != nil {
		t.Skip("could not create test audio:", err)
	}

	// Test conversions: WAV → MP3 → FLAC → OGG
	tests := []struct {
		ext  string
		name string
	}{
		{"mp3", "MP3"},
		{"flac", "FLAC"},
		{"ogg", "OGG"},
	}

	prevFile := wav
	for _, tc := range tests {
		outFile := filepath.Join(dir, "out."+tc.ext)
		cmd := exec.Command(flux, "convert", "-i", prevFile, "-o", outFile, "--engine", "ffmpeg")
		cmd.Env = append(os.Environ(), "NO_COLOR=1")
		if outb, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s conversion failed: %v\n%s", tc.name, err, outb)
		}

		stat, err := os.Stat(outFile)
		if err != nil {
			t.Fatalf("%s output file not created", tc.name)
		}

		if stat.Size() == 0 {
			t.Errorf("%s output file is empty", tc.name)
		}

		prevFile = outFile
	}
}

