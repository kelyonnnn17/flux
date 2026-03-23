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

	// Common formatting for both professional and technical
	args = append(args, "--standalone") // Produce a complete document (not a fragment)
	args = append(args, "--toc")        // Table of contents
	args = append(args, "--number-sections")
	args = append(args, "--citeproc") // Citation/bibliography support

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(outputPath), "."))

	switch f.Style {
	case "professional":
		// Professional defaults: serif fonts, proper margins, formal styling
		switch ext {
		case "pdf":
			// PDF: Times font, xelatex engine for better typography
			args = append(args, "--pdf-engine=xelatex")
			args = append(args, "-V", "fontfamily:times")
			args = append(args, "-V", "geometry:margin=1in")
			args = append(args, "-V", "linestretch=1.15")
		case "html":
			// HTML: responsive styling with good typography
			args = append(args, "-V", "css=https://cdn.jsdelivr.net/npm/water.css@2.1.1/out/dark.min.css")
			args = append(args, "-V", "classoption=twocolumn")
		case "docx":
			// DOCX: standard margins and spacing
			args = append(args, "-V", "margin-top=1in")
			args = append(args, "-V", "margin-bottom=1in")
			args = append(args, "-V", "margin-left=1in")
			args = append(args, "-V", "margin-right=1in")
		}

	case "technical":
		// Technical defaults: code clarity, monospace, syntax highlighting
		switch ext {
		case "pdf":
			// PDF: monospace font for technical content
			args = append(args, "--pdf-engine=xelatex")
			args = append(args, "-V", "fontfamily:courier")
			args = append(args, "-V", "geometry:margin=0.75in")
			args = append(args, "--highlight-style=tango")
		case "html":
			// HTML: code-friendly styling
			args = append(args, "--highlight-style=tango")
			args = append(args, "-V", "css=https://cdn.jsdelivr.net/npm/highlight.js@11.7.0/styles/atom-one-dark.min.css")
		case "docx":
			// DOCX: monospace code blocks
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
