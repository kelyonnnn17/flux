package format

import (
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
			contains: []string{"--standalone", "--toc", "--number-sections", "--citeproc"},
			notEmpty: true,
		},
		{
			name:     "professional pdf",
			style:    "professional",
			output:   "output.pdf",
			contains: []string{"--standalone", "--pdf-engine=xelatex", "fontfamily:times"},
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
			contains: []string{"--standalone", "--toc"},
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
		})
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
