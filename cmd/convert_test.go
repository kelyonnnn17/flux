package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMergeConvertArgs_PositionalOnly(t *testing.T) {
	inputs, out, err := mergeConvertArgs([]string{"file.md", "pdf"}, nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inputs) != 1 || inputs[0] != "file.md" {
		t.Fatalf("unexpected inputs: %#v", inputs)
	}
	if out != "pdf" {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestMergeConvertArgs_BackwardCompatibilityFlags(t *testing.T) {
	inputs, out, err := mergeConvertArgs(nil, []string{"file.md"}, "file.pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inputs) != 1 || inputs[0] != "file.md" {
		t.Fatalf("unexpected inputs: %#v", inputs)
	}
	if out != "file.pdf" {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestMergeConvertArgs_RejectMixedInputModes(t *testing.T) {
	_, _, err := mergeConvertArgs([]string{"file.md", "pdf"}, []string{"other.md"}, "")
	if err == nil {
		t.Fatal("expected error when mixing positional input and -i")
	}
}

func TestMergeConvertArgs_RejectMixedOutputModes(t *testing.T) {
	_, _, err := mergeConvertArgs([]string{"file.md", "pdf"}, nil, "out.pdf")
	if err == nil {
		t.Fatal("expected error when mixing positional output and -o")
	}
}

func TestMergeConvertArgs_TooManyArgs(t *testing.T) {
	_, _, err := mergeConvertArgs([]string{"a", "b", "c"}, nil, "")
	if err == nil {
		t.Fatal("expected error for too many args")
	}
}

func TestRunBatchMode_ContinueOnError(t *testing.T) {
	dir := t.TempDir()

	validFile := filepath.Join(dir, "valid.json")
	if err := os.WriteFile(validFile, []byte(`[{"key":"val"}]`), 0644); err != nil {
		t.Fatal(err)
	}
	badFile := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(badFile, []byte(`not json at all`), 0644); err != nil {
		t.Fatal(err)
	}

	// Without --continue-on-error, stops at first failure and valid.csv is never created.
	forceFlag = true
	quietFlag = true
	continueOnErrorFlag = false
	err := runBatchMode([]string{badFile, validFile}, "csv", "auto", "none", "")
	if err == nil {
		t.Fatal("expected error")
	}
	if _, statErr := os.Stat(filepath.Join(dir, "valid.csv")); statErr == nil {
		t.Fatal("valid.csv should not exist when fail-fast stops batch early")
	}

	// With --continue-on-error, valid.csv is produced despite bad.json failing.
	continueOnErrorFlag = true
	err = runBatchMode([]string{badFile, validFile}, "csv", "auto", "none", "")
	if err == nil {
		t.Fatal("expected error summarising failures")
	}
	if _, statErr := os.Stat(filepath.Join(dir, "valid.csv")); statErr != nil {
		t.Fatalf("valid.csv should exist when --continue-on-error is set: %v", statErr)
	}

	// Reset globals.
	forceFlag = false
	quietFlag = false
	continueOnErrorFlag = false
}
