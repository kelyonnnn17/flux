package engine

import (
	"fmt"
	"os"
)

// magicBytes maps formats to the byte sequences that valid files must start with.
var magicBytes = map[string][]byte{
	"pdf":  []byte("%PDF"),
	"docx": []byte("PK\x03\x04"), // DOCX is a ZIP container
	"xlsx": []byte("PK\x03\x04"),
	"pptx": []byte("PK\x03\x04"),
	"epub": []byte("PK\x03\x04"),
}

// validateOutputFile verifies that path exists, is non-empty, and (for known
// formats) starts with the expected magic bytes. It is called after every
// conversion step so silent empty-file failures are caught early.
func validateOutputFile(path, format string) error {
	st, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("output file not found: %s", path)
	}
	if st.IsDir() {
		return fmt.Errorf("output path is a directory, not a file: %s", path)
	}
	if st.Size() == 0 {
		return fmt.Errorf("output file is empty: %s", path)
	}

	magic, ok := magicBytes[format]
	if !ok {
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open output for validation: %w", err)
	}
	defer f.Close()

	header := make([]byte, len(magic))
	n, err := f.Read(header)
	if err != nil || n < len(magic) {
		return fmt.Errorf("output file too small to validate format %s: %s", format, path)
	}
	for i, b := range magic {
		if header[i] != b {
			return fmt.Errorf("output file does not look like %s (bad magic bytes): %s", format, path)
		}
	}
	return nil
}
