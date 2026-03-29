package fileutil

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v4"
)

// WriteYAMLAtomic marshals v to YAML and writes it to path using a
// temp-file-rename so that a crash mid-write cannot corrupt the original file.
func WriteYAMLAtomic(path string, v any) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("error marshaling YAML: %v", err)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(path), "tmp_ipam.*.txt")
	if err != nil {
		return fmt.Errorf("error creating temp file: %v", err)
	}

	cleanup := true
	defer func() {
		if cleanup {
			if err := os.Remove(tmpFile.Name()); err != nil {
				log.Printf("error removing temp file: %v", err)
			}
		}
	}()

	if _, err = tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("error writing to temp file: %v", err)
	}

	if err = tmpFile.Close(); err != nil {
		return fmt.Errorf("error closing temp file: %v", err)
	}

	if err = os.Rename(tmpFile.Name(), path); err != nil {
		return fmt.Errorf("error writing IPAM data: %v", err)
	}

	cleanup = false
	return nil
}
