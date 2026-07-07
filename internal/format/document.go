package format

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// typstFixLinksLua is a Pandoc Lua filter that rewrites internal link targets
// when converting to Typst via --pdf-engine=typst.
//
// Problem: Pandoc strips leading number-dot/number-space prefixes from heading
// identifiers (e.g. "## 1. Executive summary" → id "executive-summary"), but
// hand-written Markdown TOC links often use the full numbered anchor
// (#1-executive-summary). The Typst writer emits #link(<1-executive-summary>)
// which Typst cannot resolve, producing "label does not exist" errors.
//
// Fix: collect every heading id Pandoc assigns, then for each Link whose target
// is not a known heading id, strip the leading "N-" prefix and retry.
const typstFixLinksLua = `
-- typst_fix_links.lua (embedded)
-- Two-pass Pandoc filter: collect all heading identifiers first, then rewrite
-- any internal link whose target doesn't match a known heading id.
-- This handles the case where TOC links appear before headings in the document,
-- which would cause a single-pass filter to see an empty heading_ids table.

local function collect_heading_ids(blocks, ids)
  for _, block in ipairs(blocks) do
    if block.t == "Header" then
      if block.identifier and block.identifier ~= "" then
        ids[block.identifier] = true
      end
    elseif block.t == "Div" or block.t == "BlockQuote" or block.t == "Note" then
      collect_heading_ids(block.content or {}, ids)
    elseif block.t == "BulletList" or block.t == "OrderedList" then
      for _, item in ipairs(block.content or {}) do
        collect_heading_ids(item, ids)
      end
    end
  end
end

-- normalize collapses repeated hyphens (e.g. from em-dash → "--") to a single hyphen,
-- and trims any leading/trailing hyphens. This matches what Pandoc's auto_identifiers
-- extension does when it encounters special characters like em dashes.
local function normalize(id)
  return id:gsub("%-%-+", "-"):gsub("^%-+", ""):gsub("%-+$", "")
end

-- Try progressively looser matches for a link target id against the known heading ids.
-- Returns the matching heading id, or nil if no match found.
local function resolve_id(raw_id, ids)
  -- 1. Exact match (already handled by caller, but cheap to include)
  if ids[raw_id] then return raw_id end

  -- 2. Strip leading "N-" number prefix (e.g. "1-executive-summary" → "executive-summary")
  local stripped = raw_id:match("^%d+%-(.+)$")
  if stripped then
    if ids[stripped] then return stripped end
    -- 3. Also normalize double hyphens in the stripped form
    -- (e.g. "tamper-evidence--theory" → "tamper-evidence-theory")
    local norm = normalize(stripped)
    if norm ~= stripped and ids[norm] then return norm end
  end

  -- 4. Normalize without stripping (in case anchor has no number prefix but has double hyphens)
  local norm_raw = normalize(raw_id)
  if norm_raw ~= raw_id and ids[norm_raw] then return norm_raw end

  return nil
end

local function fix_links_in_blocks(blocks, ids)
  return pandoc.walk_block(
    pandoc.Div(blocks),
    {
      Link = function(el)
        local target = el.target
        if not target:match("^#") then return el end
        local raw_id = target:sub(2)
        if ids[raw_id] then return el end
        local resolved = resolve_id(raw_id, ids)
        if resolved then
          el.target = "#" .. resolved
          return el
        end
        return el
      end
    }
  ).content
end

function Pandoc(doc)
  local ids = {}
  collect_heading_ids(doc.blocks, ids)
  local new_blocks = fix_links_in_blocks(doc.blocks, ids)
  return pandoc.Pandoc(new_blocks, doc.meta)
end
`

// writeTypstFixLinksFilter writes the embedded Lua filter to a temp file and
// returns its path. The caller is responsible for removing the file.
func writeTypstFixLinksFilter() (string, error) {
	f, err := os.CreateTemp("", "flux-typst-fix-*.lua")
	if err != nil {
		return "", err
	}
	if _, err := f.WriteString(typstFixLinksLua); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", err
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

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

	if ext == "pdf" {
		// xelatex handles Unicode (em dashes, arrows, symbols) that pdflatex rejects.
		// Fallback to typst if xelatex is not available.
		if _, err := exec.LookPath("xelatex"); err == nil {
			args = append(args, "--pdf-engine=xelatex")
		} else if _, err := exec.LookPath("typst"); err == nil {
			args = append(args, "--pdf-engine=typst")
			// Apply the Lua filter that fixes numbered-heading anchor mismatches
			// (e.g. #1-executive-summary → #executive-summary) so that Typst can
			// resolve all #link() targets without "label does not exist" errors.
			if filterPath, err := writeTypstFixLinksFilter(); err == nil {
				args = append(args, "--lua-filter="+filterPath)
				// Note: the temp file persists until the process exits; pandoc reads
				// it synchronously so we cannot safely delete it before pandoc finishes.
				// The OS will reclaim it on process exit.
			}
		} else {
			args = append(args, "--pdf-engine=xelatex")
		}
	}

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
