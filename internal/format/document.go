package format

import (
	"path/filepath"
	"strings"
)

// DocumentFormatter applies sensible formatting defaults to document conversions.
type DocumentFormatter struct {
	Style string // "professional", "technical", or "none"
}

// NewDocumentFormatter creates a formatter with the given style preset.
// Style can be "professional", "technical", or "none" (no formatting).
func NewDocumentFormatter(style string) *DocumentFormatter {
	if style == "" {
		style = "professional" // default
	}
	return &DocumentFormatter{Style: style}
}

// PandocArgs returns Pandoc CLI arguments based on the formatter style and output format.
func (f *DocumentFormatter) PandocArgs(outputPath string) []string {
	if f.Style == "none" {
		return []string{}
	}

	args := []string{}

	// Keep defaults non-intrusive: preserve source structure and avoid synthetic sections/TOC.
	args = append(args, "--standalone")
	args = append(args, "--citeproc")

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(outputPath), "."))

	switch f.Style {
	case "professional":
		// Professional keeps minimal defaults so output reflects source intent.
		switch ext {
		case "pdf", "html", "docx", "odt", "md", "tex", "epub", "rst":
			// No injected layout/theme flags by default.
		}

	case "technical":
		// Technical may adjust code rendering but does not inject TOC/numbering/layout.
		switch ext {
		case "pdf":
			args = append(args, "--highlight-style=tango")
		case "html":
			args = append(args, "--highlight-style=tango")
		case "docx":
			args = append(args, "--highlight-style=tango")
		}
	}

	return args
}

// IsDocumentFormat checks if the output format should be formatted.
func IsDocumentFormat(ext string) bool {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))
	documentExts := map[string]bool{
		"pdf":  true,
		"html": true,
		"docx": true,
		"odt":  true,
		"md":   true,
		"tex":  true,
		"epub": true,
		"rst":  true,
	}
	return documentExts[ext]
}
