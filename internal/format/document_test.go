package format

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocumentFormatter_PandocArgs(t *testing.T) {
	tests := []struct {
		name     string
		style    string
		output   string
		contains []string
		notEmpty bool
	}{
		{
			name:     "professional html",
			style:    "professional",
			output:   "output.html",
			contains: []string{"--standalone", "--citeproc"},
			notEmpty: true,
		},
		{
			name:     "professional pdf",
			style:    "professional",
			output:   "output.pdf",
			contains: []string{"--standalone", "--citeproc"},
			notEmpty: true,
		},
		{
			name:     "technical html",
			style:    "technical",
			output:   "output.html",
			contains: []string{"--standalone", "--highlight-style=tango"},
			notEmpty: true,
		},
		{
			name:     "developer pdf",
			style:    "developer",
			output:   "output.pdf",
			contains: []string{"--standalone", "--highlight-style=zenburn", "--variable=geometry:margin=1in", "--variable=linestretch:1.2"},
			notEmpty: true,
		},
		{
			name:     "developer docx",
			style:    "developer",
			output:   "output.docx",
			contains: []string{"--standalone", "--highlight-style=zenburn", "--metadata=title:Developer Document"},
			notEmpty: true,
		},
		{
			name:     "none style",
			style:    "none",
			output:   "output.html",
			contains: []string{},
			notEmpty: false,
		},
		{
			name:     "default style",
			style:    "",
			output:   "output.html",
			contains: []string{"--standalone", "--citeproc"},
			notEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewDocumentFormatter(tt.style)
			args := f.PandocArgs(tt.output)

			if tt.notEmpty && len(args) == 0 {
				t.Errorf("expected non-empty args for style %s, got empty", tt.style)
			}
			if !tt.notEmpty && len(args) > 0 {
				t.Errorf("expected empty args for style %s, got %v", tt.style, args)
			}

			argsStr := strings.Join(args, " ")
			for _, want := range tt.contains {
				if !strings.Contains(argsStr, want) {
					t.Errorf("expected args to contain %q, got %v", want, args)
				}
			}

			if strings.Contains(argsStr, "--toc") {
				t.Errorf("did not expect auto TOC args, got %v", args)
			}
			if strings.Contains(argsStr, "--number-sections") {
				t.Errorf("did not expect auto section numbering args, got %v", args)
			}
		})
	}
}

func TestDocumentFormatter_PandocArgsWithContext_DeveloperMarkdownWrap(t *testing.T) {
	f := NewDocumentFormatter("developer")
	args := f.PandocArgsWithContext("guide.md", "guide.pdf", "")
	argsStr := strings.Join(args, " ")
	if !strings.Contains(argsStr, "--wrap=preserve") {
		t.Fatalf("expected developer markdown conversion to preserve wrapping, got %v", args)
	}
}

func TestIsDocumentFormat(t *testing.T) {
	tests := []struct {
		ext  string
		want bool
	}{
		{"pdf", true},
		{"html", true},
		{"md", true},
		{"docx", true},
		{"rst", true},
		{"jpg", false},
		{"png", false},
		{"mp3", false},
		{"csv", false},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			got := IsDocumentFormat(tt.ext)
			if got != tt.want {
				t.Errorf("IsDocumentFormat(%s) = %v, want %v", tt.ext, got, tt.want)
			}
		})
	}
}

func TestDocumentFormatter_PandocArgsWithContext_DOCXReference(t *testing.T) {
	t.Run("auto reference from input docx", func(t *testing.T) {
		dir := t.TempDir()
		in := filepath.Join(dir, "in.docx")
		if err := os.WriteFile(in, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}

		f := NewDocumentFormatter("professional")
		args := f.PandocArgsWithContext(in, filepath.Join(dir, "out.docx"), "")
		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "--reference-doc="+in) {
			t.Fatalf("expected auto reference-doc from input docx, got %v", args)
		}
	})

	t.Run("explicit reference doc", func(t *testing.T) {
		dir := t.TempDir()
		reference := filepath.Join(dir, "reference.docx")
		if err := os.WriteFile(reference, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}

		f := NewDocumentFormatter("professional")
		args := f.PandocArgsWithContext("", filepath.Join(dir, "out.docx"), reference)
		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "--reference-doc="+reference) {
			t.Fatalf("expected explicit reference-doc, got %v", args)
		}
	})

	t.Run("ignore missing reference doc", func(t *testing.T) {
		dir := t.TempDir()
		missing := filepath.Join(dir, "missing.docx")

		f := NewDocumentFormatter("professional")
		args := f.PandocArgsWithContext("", filepath.Join(dir, "out.docx"), missing)
		joined := strings.Join(args, " ")
		if strings.Contains(joined, "--reference-doc=") {
			t.Fatalf("did not expect reference-doc for missing file, got %v", args)
		}
	})
}
