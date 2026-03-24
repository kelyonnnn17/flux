package flux

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvert_ExplicitEngineValidation(t *testing.T) {
	t.Run("pandoc cannot convert image->image", func(t *testing.T) {
		err := Convert("input.jpg", "output.png", ConvertOptions{Engine: "pandoc"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "engine pandoc cannot convert")
		assert.Contains(t, err.Error(), "imagemagick")
	})

	t.Run("ffmpeg cannot convert data->data", func(t *testing.T) {
		err := Convert("input.json", "output.yaml", ConvertOptions{Engine: "ffmpeg"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "engine ffmpeg cannot convert")
		assert.Contains(t, err.Error(), "data")
	})

	t.Run("pandoc cannot handle pdf source end-to-end", func(t *testing.T) {
		err := Convert("input.pdf", "output.docx", ConvertOptions{Engine: "pandoc"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--engine auto")
	})
}

func TestConvert_DataEngine_Succeeds(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "in.json")
	dst := filepath.Join(dir, "out.yaml")

	require.NoError(t, os.WriteFile(src, []byte(`{"a":1,"b":"x"}`), 0644))

	err := Convert(src, dst, ConvertOptions{Engine: "data"})
	require.NoError(t, err)

	out, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Contains(t, string(out), "a:")
	assert.Contains(t, string(out), "b:")
}

