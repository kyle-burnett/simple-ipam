package add

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/kyle-burnett/simple-ipam/internal/models"
	"gopkg.in/yaml.v3"
)

func Test_AddSubnet(t *testing.T) {
	cmdSubnet := AddCmd
	cmdSubnet.SetArgs([]string{
		fmt.Sprintf("-d=%s", "test subnet"),
		fmt.Sprintf("-s=%s", "10.10.0.0/25"),
		fmt.Sprintf("-f=%s", "testAdd.yaml"),
	})
	testFile := createTestFile("testAdd.yaml")
	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Fatal(err)
		}
	}()
	expectedYAML := models.IPAM{
		Subnets: map[string]models.Subnets{
			"10.10.0.0/20": {
				Description: "test subnet",
				Tags:        []string{"tag_1", "tag_2"},
				Subnets: map[string]models.Subnets{
					"10.10.0.0/24": {
						Description: "test subnet",
						Tags:        []string{"tag_1", "tag_2"},
						Subnets: map[string]models.Subnets{
							"10.10.0.0/25": {
								Description: "test subnet",
								Tags:        []string{},
								Subnets:     map[string]models.Subnets{},
							},
						},
					},
				},
			},
		},
	}

	expectedYamlData, err := yaml.Marshal(&expectedYAML)
	if err != nil {
		fmt.Printf("Error while Marshaling. %v", err)
	}

	err = cmdSubnet.Execute()
	if err != nil {
		t.Fatal(err)
	}

	ipamFile, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Error reading YAML file: %v", err)
	}

	var expectedData interface{}
	if err := yaml.Unmarshal([]byte(expectedYamlData), &expectedData); err != nil {
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

func Test_AddSupernet(t *testing.T) {
	cmdSupernet := AddCmd
	cmdSupernet.SetArgs([]string{
		fmt.Sprintf("-d=%s", "test subnet"),
		fmt.Sprintf("-s=%s", "10.10.0.0/22"),
		fmt.Sprintf("-f=%s", "testSupernet.yaml"),
	})
	testFile := createTestFile("testSupernet.yaml")
	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Fatal(err)
		}
	}()
	expectedYAML := models.IPAM{
		Subnets: map[string]models.Subnets{
			"10.10.0.0/20": {
				Description: "test subnet",
				Tags:        []string{"tag_1", "tag_2"},
				Subnets: map[string]models.Subnets{
					"10.10.0.0/22": {
						Description: "test subnet",
						Tags:        []string{},
						Subnets: map[string]models.Subnets{
							"10.10.0.0/24": {
								Description: "test subnet",
								Tags:        []string{"tag_1", "tag_2"},
								Subnets:     map[string]models.Subnets{},
							},
						},
					},
				},
			},
		},
	}

	expectedYamlData, err := yaml.Marshal(&expectedYAML)
	if err != nil {
		fmt.Printf("Error while Marshaling. %v", err)
	}

	err = cmdSupernet.Execute()
	if err != nil {
		t.Fatal(err)
	}

	ipamFile, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Error reading YAML file: %v", err)
	}

	var expectedData interface{}
	if err := yaml.Unmarshal([]byte(expectedYamlData), &expectedData); err != nil {
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

func createTestFile(fileName string) string {
	ipamData := models.IPAM{
		Subnets: map[string]models.Subnets{
			"10.10.0.0/20": {
				Description: "test subnet",
				Tags:        []string{"tag_1", "tag_2"},
				Subnets: map[string]models.Subnets{
					"10.10.0.0/24": {
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
