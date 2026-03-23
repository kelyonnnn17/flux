package cmd

import (
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
