package cmd

import (
	"reflect"
	"testing"
)

func TestRewriteArgsForShortcut(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "explicit command unchanged",
			in:   []string{"flux", "convert", "a.md", "pdf"},
			want: []string{"flux", "convert", "a.md", "pdf"},
		},
		{
			name: "shortcut inserts convert",
			in:   []string{"flux", "a.md", "pdf"},
			want: []string{"flux", "convert", "a.md", "pdf"},
		},
		{
			name: "shortcut with persistent flag",
			in:   []string{"flux", "--engine", "pandoc", "a.md", "pdf"},
			want: []string{"flux", "--engine", "pandoc", "convert", "a.md", "pdf"},
		},
		{
			name: "known short command unchanged",
			in:   []string{"flux", "d"},
			want: []string{"flux", "d"},
		},
		{
			name: "single argument unchanged",
			in:   []string{"flux", "a.md"},
			want: []string{"flux", "a.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rewriteArgsForShortcut(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("rewriteArgsForShortcut() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
