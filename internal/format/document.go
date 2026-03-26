package format

import (
	"os"
	"path/filepath"
	"strings"
)

// DocumentFormatter applies sensible formatting defaults to document conversions.
type DocumentFormatter struct {
	Style string // "professional", "technical", "developer", or "none"
}

// NewDocumentFormatter creates a formatter with the given style preset.
// Style can be "professional", "technical", "developer", or "none" (no formatting).
func NewDocumentFormatter(style string) *DocumentFormatter {
	if style == "" {
		style = "professional" // default
	}
	return &DocumentFormatter{Style: style}
}

// PandocArgs returns Pandoc CLI arguments based on the formatter style and output format.
func (f *DocumentFormatter) PandocArgs(outputPath string) []string {
	return f.PandocArgsWithContext("", outputPath, "")
}

// PandocArgsWithContext returns Pandoc CLI arguments based on style and paths.
// For DOCX output, it can attach a reference DOCX to preserve styles.
func (f *DocumentFormatter) PandocArgsWithContext(inputPath, outputPath, referenceDocPath string) []string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(outputPath), "."))
	inputExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(inputPath), "."))

	args := []string{}

	if f.Style != "none" {
		// Keep defaults non-intrusive: preserve source structure and avoid synthetic sections/TOC.
		args = append(args, "--standalone")
		args = append(args, "--citeproc")

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
			case "pdf", "html", "docx":
				args = append(args, "--highlight-style=tango")
			}

		case "developer":
			// Developer preset optimizes readability of markdown docs with code and tables.
			switch ext {
			case "pdf":
				args = append(args,
					"--highlight-style=zenburn",
					"--variable=geometry:margin=1in",
					"--variable=fontsize:11pt",
					"--variable=linestretch:1.2",
				)
			case "docx":
				args = append(args,
					"--highlight-style=zenburn",
					"--metadata=title:Developer Document",
				)
			case "html":
				args = append(args, "--highlight-style=zenburn")
			}

			if inputExt == "md" || inputExt == "markdown" {
				args = append(args, "--wrap=preserve")
			}
		}
	}

	if ext == "docx" {
		refDoc := strings.TrimSpace(referenceDocPath)
		if refDoc == "" {
			inputExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(inputPath), "."))
			if inputExt == "docx" {
				refDoc = inputPath
			}
		}

		if refDoc != "" {
			if st, err := os.Stat(refDoc); err == nil && !st.IsDir() {
				args = append(args, "--reference-doc="+refDoc)
			}
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
