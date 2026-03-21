package data

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvert_JSONToYAML(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "in.json")
	dst := filepath.Join(dir, "out.yaml")
	require.NoError(t, os.WriteFile(src, []byte(`{"a":1,"b":"x"}`), 0644))

	err := Convert(src, dst, FormatJSON, FormatYAML)
	require.NoError(t, err)

	out, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Contains(t, string(out), "a:")
	assert.Contains(t, string(out), "b:")
}

func TestConvert_JSONToCSV(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "in.json")
	dst := filepath.Join(dir, "out.csv")
	jsonData := `[{"name":"alice","age":"30"},{"name":"bob","age":"25"}]`
	require.NoError(t, os.WriteFile(src, []byte(jsonData), 0644))

	err := Convert(src, dst, FormatJSON, FormatCSV)
	require.NoError(t, err)

	out, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Contains(t, string(out), "name")
	assert.Contains(t, string(out), "alice")
	assert.Contains(t, string(out), "bob")
}

func TestConvert_CSVToJSON(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "in.csv")
	dst := filepath.Join(dir, "out.json")
	csvData := "name,age\nalice,30\nbob,25\n"
	require.NoError(t, os.WriteFile(src, []byte(csvData), 0644))

	err := Convert(src, dst, FormatCSV, FormatJSON)
	require.NoError(t, err)

	out, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Contains(t, string(out), "alice")
	assert.Contains(t, string(out), "bob")
	assert.Contains(t, string(out), "30")
}

func TestConvert_CSVToYAML(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "in.csv")
	dst := filepath.Join(dir, "out.yaml")
	csvData := "x,y\n1,2\n3,4\n"
	require.NoError(t, os.WriteFile(src, []byte(csvData), 0644))

	err := Convert(src, dst, FormatCSV, FormatYAML)
	require.NoError(t, err)

	out, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Contains(t, string(out), "x")
	assert.Contains(t, string(out), "y")
}

func TestConvert_YAMLToJSON(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "in.yaml")
	dst := filepath.Join(dir, "out.json")
	yamlData := "key: value\nnum: 42\n"
	require.NoError(t, os.WriteFile(src, []byte(yamlData), 0644))

	err := Convert(src, dst, FormatYAML, FormatJSON)
	require.NoError(t, err)

	out, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Contains(t, string(out), "key")
	assert.Contains(t, string(out), "value")
	assert.Contains(t, string(out), "42")
}

func TestConvert_SameFormatError(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "in.json")
	dst := filepath.Join(dir, "out.json")
	require.NoError(t, os.WriteFile(src, []byte("{}"), 0644))

	err := Convert(src, dst, FormatJSON, FormatJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "same")
}

func TestConvert_InferFromPath(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "data.json")
	dst := filepath.Join(dir, "data.yaml")
	require.NoError(t, os.WriteFile(src, []byte(`{"x":1}`), 0644))

	err := Convert(src, dst, "", "")
	require.NoError(t, err)

	out, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Contains(t, string(out), "x:")
}

func TestFormatFromExt(t *testing.T) {
	assert.Equal(t, FormatJSON, FormatFromExt("file.json"))
	assert.Equal(t, FormatYAML, FormatFromExt("file.yml"))
	assert.Equal(t, FormatYAML, FormatFromExt("file.yaml"))
	assert.Equal(t, FormatCSV, FormatFromExt("file.csv"))
	assert.Equal(t, FormatTOML, FormatFromExt("file.toml"))
	assert.Equal(t, "", FormatFromExt("file.unknown"))
}

func TestIsDataFormat(t *testing.T) {
	assert.True(t, IsDataFormat(".json"))
	assert.True(t, IsDataFormat("json"))
	assert.True(t, IsDataFormat(".csv"))
	assert.True(t, IsDataFormat(".yaml"))
	assert.True(t, IsDataFormat(".yml"))
	assert.True(t, IsDataFormat(".toml"))
	assert.False(t, IsDataFormat(".txt"))
	assert.False(t, IsDataFormat(".pdf"))
}

func TestConvert_TSVToJSON(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "in.tsv")
	dst := filepath.Join(dir, "out.json")
	tsvData := "a\tb\n1\t2\n3\t4\n"
	require.NoError(t, os.WriteFile(src, []byte(tsvData), 0644))

	err := Convert(src, dst, FormatTSV, FormatJSON)
	require.NoError(t, err)

	out, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Contains(t, string(out), "a")
	assert.Contains(t, string(out), "b")
}

func TestConvert_TOMLToJSON(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "in.toml")
	dst := filepath.Join(dir, "out.json")
	tomlData := "title = \"test\"\ncount = 42\n"
	require.NoError(t, os.WriteFile(src, []byte(tomlData), 0644))

	err := Convert(src, dst, FormatTOML, FormatJSON)
	require.NoError(t, err)

	out, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Contains(t, string(out), "title")
	assert.Contains(t, string(out), "test")
	assert.Contains(t, string(out), "42")
}
