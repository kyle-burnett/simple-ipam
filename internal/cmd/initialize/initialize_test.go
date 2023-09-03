package initialize

import (
	"fmt"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func Test_InitCommand(t *testing.T) {
	cmd := InitCmd
	cmd.SetArgs([]string{
		fmt.Sprintf("-d=%s", "test"),
		fmt.Sprintf("-f=%s", "test"),
	})
	cmd.Execute()
	ipamFile, err := os.ReadFile("test.yaml")
	if err != nil {
		t.Fatalf("Error reading YAML file: %v", err)
	}
	expectedYAML := `
ipam:
  description: test
  prefixes: {}
`

	var expectedData interface{}
	if err := yaml.Unmarshal([]byte(expectedYAML), &expectedData); err != nil {
		os.Remove("test.yaml")
		t.Fatalf("Error unmarshaling expected YAML: %v", err)
	}

	var actualData interface{}
	if err := yaml.Unmarshal(ipamFile, &actualData); err != nil {
		os.Remove("test.yaml")
		t.Fatalf("Error unmarshaling actual YAML: %v", err)
	}

	if fmt.Sprintf("%v", actualData) != fmt.Sprintf("%v", expectedData) {
		t.Errorf("YAML content does not match expected content")
	}
	os.Remove("test.yaml")
}
