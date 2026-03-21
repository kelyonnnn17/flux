package data

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// Format names used for --from and --to flags.
const (
	FormatCSV  = "csv"
	FormatTSV  = "tsv"
	FormatJSON = "json"
	FormatYAML = "yaml"
	FormatTOML = "toml"
)

// IsDataFormat returns true if ext is a supported data format.
func IsDataFormat(ext string) bool {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))
	switch ext {
	case "csv", "tsv", "json", "yaml", "yml", "toml":
		return true
	}
	return false
}

// FormatFromExt returns the canonical format name from a file extension.
func FormatFromExt(path string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	switch ext {
	case "yml":
		return FormatYAML
	case "csv", "tsv", "json", "yaml", "toml":
		return ext
	}
	return ""
}

// Convert converts between data formats. fromFormat and toFormat can be empty
// to infer from file extensions.
func Convert(src, dst, fromFormat, toFormat string) error {
	if fromFormat == "" {
		fromFormat = FormatFromExt(src)
	}
	if toFormat == "" {
		toFormat = FormatFromExt(dst)
	}
	if fromFormat == "" || toFormat == "" {
		return fmt.Errorf("cannot infer format: use --from and --to to specify")
	}
	if fromFormat == toFormat {
		return fmt.Errorf("source and target format are the same (%s)", fromFormat)
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}

	var parsed interface{}
	if err := unmarshal(data, fromFormat, &parsed); err != nil {
		return fmt.Errorf("parse %s as %s: %w", src, fromFormat, err)
	}

	out, err := marshal(parsed, toFormat)
	if err != nil {
		return fmt.Errorf("encode as %s: %w", toFormat, err)
	}

	if err := os.WriteFile(dst, out, 0644); err != nil {
		return fmt.Errorf("write %s: %w", dst, err)
	}
	return nil
}

// ConvertStream converts from reader to writer (for pipe support).
func ConvertStream(r io.Reader, w io.Writer, fromFormat, toFormat string) error {
	if fromFormat == "" || toFormat == "" {
		return fmt.Errorf("--from and --to required for pipe mode")
	}
	if fromFormat == toFormat {
		return fmt.Errorf("source and target format are the same (%s)", fromFormat)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	var parsed interface{}
	if err := unmarshal(data, fromFormat, &parsed); err != nil {
		return fmt.Errorf("parse as %s: %w", fromFormat, err)
	}
	out, err := marshal(parsed, toFormat)
	if err != nil {
		return fmt.Errorf("encode as %s: %w", toFormat, err)
	}
	if _, err := w.Write(out); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

func unmarshal(data []byte, format string, v *interface{}) error {
	switch format {
	case FormatCSV, FormatTSV:
		return unmarshalCSV(data, format == FormatTSV, v)
	case FormatJSON:
		return json.Unmarshal(data, v)
	case FormatYAML:
		return yaml.Unmarshal(data, v)
	case FormatTOML:
		_, err := toml.Decode(string(data), v)
		return err
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func unmarshalCSV(data []byte, isTSV bool, v *interface{}) error {
	r := csv.NewReader(bytes.NewReader(data))
	if isTSV {
		r.Comma = '\t'
	}
	records, err := r.ReadAll()
	if err != nil {
		return err
	}
	if len(records) == 0 {
		*v = []interface{}{}
		return nil
	}
	headers := records[0]
	rows := make([]map[string]interface{}, 0, len(records)-1)
	for i := 1; i < len(records); i++ {
		row := make(map[string]interface{})
		for j, h := range headers {
			val := ""
			if j < len(records[i]) {
				val = records[i][j]
			}
			row[h] = val
		}
		rows = append(rows, row)
	}
	*v = rows
	return nil
}

func marshal(v interface{}, format string) ([]byte, error) {
	switch format {
	case FormatCSV, FormatTSV:
		return marshalCSV(v, format == FormatTSV)
	case FormatJSON:
		return json.MarshalIndent(v, "", "  ")
	case FormatYAML:
		return yaml.Marshal(v)
	case FormatTOML:
		var buf bytes.Buffer
		err := toml.NewEncoder(&buf).Encode(v)
		return buf.Bytes(), err
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func marshalCSV(v interface{}, isTSV bool) ([]byte, error) {
	var rows []map[string]interface{}
	switch x := v.(type) {
	case []map[string]interface{}:
		rows = x
	case []interface{}:
		for _, item := range x {
			if m, ok := item.(map[string]interface{}); ok {
				rows = append(rows, m)
			} else {
				return nil, fmt.Errorf("CSV/TSV output requires array of objects")
			}
		}
	case map[string]interface{}:
		rows = []map[string]interface{}{x}
	default:
		return nil, fmt.Errorf("CSV/TSV output requires array of objects")
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if isTSV {
		w.Comma = '\t'
	}
	if len(rows) == 0 {
		return buf.Bytes(), nil
	}
	headers := headersFromRows(rows)
	if err := w.Write(headers); err != nil {
		return nil, err
	}
	for _, row := range rows {
		record := make([]string, len(headers))
		for i, h := range headers {
			if x, ok := row[h]; ok {
				record[i] = fmt.Sprintf("%v", x)
			}
		}
		if err := w.Write(record); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func headersFromRows(rows []map[string]interface{}) []string {
	seen := make(map[string]bool)
	for _, row := range rows {
		for k := range row {
			seen[k] = true
		}
	}
	var headers []string
	if len(rows) > 0 {
		for k := range rows[0] {
			headers = append(headers, k)
		}
		for k := range seen {
			if !contains(headers, k) {
				headers = append(headers, k)
			}
		}
	}
	return headers
}

func contains(s []string, x string) bool {
	for _, v := range s {
		if v == x {
			return true
		}
	}
	return false
}
