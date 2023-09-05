package delete

import (
	"fmt"
	"log"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func Test_Delete(t *testing.T) {
	cmdDelete := DeleteCmd
	cmdDelete.SetArgs([]string{
		fmt.Sprintf("-c=%s", "10.10.0.0/24"),
		fmt.Sprintf("-i=%s", "testSubnet.yaml"),
	})
	testFile := createTestFile("testSubnet.yaml")
	expectedYAML := IPAM{
		IPAM: map[string]interface{}{
			"description": "test",
			"prefixes": map[string]interface{}{
				"10.10.0.0/20": map[string]interface{}{
					"cidr_tags":   []string{"tag_1", "tag_2"},
					"description": "test cidr",
					"subnets":     map[string]interface{}{},
				},
			},
		},
	}
	expectedYamlData, err := yaml.Marshal(&expectedYAML)

	if err != nil {
		fmt.Printf("Error while Marshaling. %v", err)
	}
	cmdDelete.Execute()
	ipamFile, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Error reading YAML file: %v", err)
	}

	var expectedData interface{}
	if err := yaml.Unmarshal([]byte(expectedYamlData), &expectedData); err != nil {
		os.Remove(testFile)
		t.Fatalf("Error unmarshaling expected YAML: %v", err)
	}

	var actualData interface{}
	if err := yaml.Unmarshal(ipamFile, &actualData); err != nil {
		os.Remove(testFile)
		t.Fatalf("Error unmarshaling actual YAML: %v", err)
	}

	if fmt.Sprintf("%v", actualData) != fmt.Sprintf("%v", expectedData) {
		os.Remove(testFile)
		t.Errorf("YAML content does not match expected content")
	}
	os.Remove(testFile)
}

func Test_DeleteForce(t *testing.T) {
	cmdDeleteSubnetFail := DeleteCmd
	cmdDeleteSubnetFail.SetArgs([]string{
		fmt.Sprintf("-c=%s", "10.10.0.0/20"),
		fmt.Sprintf("-i=%s", "testSubnet.yaml"),
		"-f",
	})
	testFile := createTestFile("testSubnet.yaml")
	expectedYAML := IPAM{
		IPAM: map[string]interface{}{
			"description": "test",
			"prefixes":    map[string]interface{}{},
		},
	}
	expectedYamlData, err := yaml.Marshal(&expectedYAML)

	if err != nil {
		fmt.Printf("Error while Marshaling. %v", err)
	}
	cmdDeleteSubnetFail.Execute()

	ipamFile, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Error reading YAML file: %v", err)
	}

	var expectedData interface{}
	if err := yaml.Unmarshal([]byte(expectedYamlData), &expectedData); err != nil {
		os.Remove(testFile)
		t.Fatalf("Error unmarshaling expected YAML: %v", err)
	}

	var actualData interface{}
	if err := yaml.Unmarshal(ipamFile, &actualData); err != nil {
		os.Remove(testFile)
		t.Fatalf("Error unmarshaling actual YAML: %v", err)
	}

	if fmt.Sprintf("%v", actualData) != fmt.Sprintf("%v", expectedData) {
		os.Remove(testFile)
		t.Errorf("YAML content does not match expected content")
	}
	os.Remove(testFile)
}

func createTestFile(fileName string) string {
	ipamData := IPAM{
		IPAM: map[string]interface{}{
			"description": "test",
			"prefixes": map[string]interface{}{
				"10.10.0.0/20": map[string]interface{}{
					"cidr_tags":   []string{"tag_1", "tag_2"},
					"description": "test cidr",
					"subnets": map[string]interface{}{
						"10.10.0.0/24": map[string]interface{}{
							"cidr_tags":   []string{"tag_1", "tag_2"},
							"description": "test cidr",
							"subnets":     map[string]interface{}{},
						},
					},
				},
			},
		},
	}
	yamlData, err := yaml.Marshal(&ipamData)
	if err != nil {
		fmt.Printf("Error while Marshaling. %v", err)
	}

	if _, err = os.Stat(fileName); err == nil {
		log.Fatal("File already exists")
	}

	err = os.WriteFile(fileName, yamlData, 0644)
	if err != nil {
		panic("Unable to write data into the file")
	}

	return fileName
}
