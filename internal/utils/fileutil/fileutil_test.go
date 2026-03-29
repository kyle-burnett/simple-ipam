package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v4"
)

func Test_WriteYAMLAtomic_HappyPath(t *testing.T) {
	type payload struct {
		Name  string `yaml:"name"`
		Value int    `yaml:"value"`
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "out.yaml")

	input := payload{Name: "test", Value: 42}
	if err := WriteYAMLAtomic(path, &input); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("unexpected error reading file: %v", err)
	}

	var got payload
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatalf("unexpected error unmarshaling: %v", err)
	}

	if got != input {
		t.Errorf("got %+v, want %+v", got, input)
	}
}

func Test_WriteYAMLAtomic_BadDirectory(t *testing.T) {
	path := "/tmp/nonexistent-simple-ipam-dir/out.yaml"

	err := WriteYAMLAtomic(path, map[string]string{"k": "v"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "error creating temp file") {
		t.Errorf("unexpected error: %v", err)
	}
}
