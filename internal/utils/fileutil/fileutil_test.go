package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func Test_WriteYAMLAtomic_HappyPath(t *testing.T) {
	type payload struct {
		Name  string `yaml:"name"`
		Value int    `yaml:"value"`
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "out.yaml")

	input := payload{Name: "test", Value: 42}
	require.NoError(t, WriteYAMLAtomic(path, &input))

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var got payload
	require.NoError(t, yaml.Unmarshal(data, &got))
	assert.Equal(t, input, got)
}

func Test_WriteYAMLAtomic_BadDirectory(t *testing.T) {
	path := "/tmp/nonexistent-simple-ipam-dir/out.yaml"

	err := WriteYAMLAtomic(path, map[string]string{"k": "v"})
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "error creating temp file"), "unexpected error: %v", err)
}
