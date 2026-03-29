package initialize

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/kyle-burnett/simple-ipam/internal/models"
)

func Test_InitCommand(t *testing.T) {
	if err := Initialize("test", "test"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.Remove("test.yaml")

	ipamFile, err := os.ReadFile("test.yaml")
	if err != nil {
		t.Fatalf("unexpected error reading file: %v", err)
	}

	expectedYAML := models.IPAM{
		Description: "test",
		Subnets:     map[string]models.Subnets{},
	}
	want, err := yaml.Marshal(&expectedYAML)
	if err != nil {
		t.Fatalf("unexpected error marshaling expected YAML: %v", err)
	}

	if string(ipamFile) != string(want) {
		t.Errorf("got:\n%s\nwant:\n%s", ipamFile, want)
	}
}

func Test_InitCommand_FileAlreadyExists(t *testing.T) {
	if err := Initialize("test", "test"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.Remove("test.yaml")

	wantErr := "IPAM file test.yaml already exists"
	err := Initialize("test", "test")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != wantErr {
		t.Errorf("got error %q, want %q", err.Error(), wantErr)
	}
}
