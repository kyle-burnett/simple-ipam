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

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := os.Remove("test.yaml")
		if err != nil {
			t.Fatal(err)
		}
	}()

	ipamFile, err := os.ReadFile("test.yaml")
	if err != nil {
		t.Fatalf("Error reading YAML file: %v", err)
	}

	expectedYAML := `
  description: test
  subnets: {}
`

	var expectedData interface{}
	if err := yaml.Unmarshal([]byte(expectedYAML), &expectedData); err != nil {
		t.Fatalf("Error unmarshaling expected YAML: %v", err)
	}

	var actualData interface{}
	if err := yaml.Unmarshal(ipamFile, &actualData); err != nil {
		t.Fatalf("Error unmarshaling actual YAML: %v", err)
	}

	if fmt.Sprintf("%v", actualData) != fmt.Sprintf("%v", expectedData) {
		t.Errorf("YAML content does not match expected content")
	}
}
