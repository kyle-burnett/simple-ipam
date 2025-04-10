package delete

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/kyle-burnett/simple-ipam/internal/models"
	"gopkg.in/yaml.v3"
)

func Test_Delete(t *testing.T) {
	cmdDelete := DeleteCmd
	cmdDelete.SetArgs([]string{
		fmt.Sprintf("-s=%s", "10.10.0.0/24"),
		fmt.Sprintf("-f=%s", "testSubnet.yaml"),
	})
	testFile := createTestFile("testSubnet.yaml")
	expectedYAML := models.IPAM{
		Subnets: map[string]models.Subnets{
			"10.10.0.0/20": {
				Description: "test subnet",
				Tags:        []string{"tag_1", "tag_2"},
				Subnets:     map[string]models.Subnets{},
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
		fmt.Sprintf("-s=%s", "10.10.0.0/20"),
		fmt.Sprintf("-f=%s", "testSubnet.yaml"),
		"-r",
	})
	testFile := createTestFile("testSubnet.yaml")
	expectedYAML := models.IPAM{
		Subnets: map[string]models.Subnets{},
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
	ipamData := models.IPAM{
		Subnets: map[string]models.Subnets{
			"10.10.0.0/20": {
				Description: "test subnet",
				Tags:        []string{"tag_1", "tag_2"},
				Subnets: map[string]models.Subnets{
					"10.10.0.0/24": models.Subnets{
						Description: "test subnet",
						Tags:        []string{"tag_1", "tag_2"},
						Subnets:     map[string]models.Subnets{},
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
