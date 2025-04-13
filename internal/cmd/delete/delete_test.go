package delete

import (
	"fmt"
	"os"
	"testing"

	"github.com/kyle-burnett/simple-ipam/internal/models"
	"github.com/kyle-burnett/simple-ipam/internal/utils/testutils"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func Test_Delete(t *testing.T) {
	cmdDelete := DeleteCmd
	cmdDelete.SetArgs([]string{
		fmt.Sprintf("-s=%s", "10.10.0.0/24"),
		fmt.Sprintf("-f=%s", "testDelete.yaml"),
	})

	testFile, err := testutils.CreateTestFile("testDelete.yaml")
	if err != nil {
		t.Errorf("Error creating test file YAML: %v", err)
	}
	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error(err)
		}
	}()
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

	err = cmdDelete.Execute()
	if err != nil {
		t.Error(err)
	}

	ipamFile, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("Error reading YAML file: %v", err)
	}

	var expectedData interface{}
	if err := yaml.Unmarshal([]byte(expectedYamlData), &expectedData); err != nil {
		t.Errorf("Error unmarshaling expected YAML: %v", err)
	}

	var actualData interface{}
	if err := yaml.Unmarshal(ipamFile, &actualData); err != nil {
		t.Errorf("Error unmarshaling actual YAML: %v", err)
	}

	if fmt.Sprintf("%v", actualData) != fmt.Sprintf("%v", expectedData) {
		t.Errorf("YAML content does not match expected content")
	}
}

func Test_DeleteRecursive(t *testing.T) {
	cmdDeleteRecursive := DeleteCmd
	cmdDeleteRecursive.SetArgs([]string{
		fmt.Sprintf("-s=%s", "10.10.0.0/20"),
		fmt.Sprintf("-f=%s", "testDeleteRecursive.yaml"),
		"-r",
	})

	testFile, err := testutils.CreateTestFile("testDeleteRecursive.yaml")
	if err != nil {
		t.Errorf("Error creating test file YAML: %v", err)
	}

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error(err)
		}
	}()
	expectedYAML := models.IPAM{
		Subnets: map[string]models.Subnets{},
	}

	expectedYamlData, err := yaml.Marshal(&expectedYAML)
	if err != nil {
		fmt.Printf("Error while Marshaling. %v", err)
	}

	err = cmdDeleteRecursive.Execute()
	if err != nil {
		t.Error(err)
	}

	ipamFile, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("Error reading YAML file: %v", err)
	}

	var expectedData interface{}
	if err := yaml.Unmarshal([]byte(expectedYamlData), &expectedData); err != nil {
		t.Errorf("Error unmarshaling expected YAML: %v", err)
	}

	var actualData interface{}
	if err := yaml.Unmarshal(ipamFile, &actualData); err != nil {
		t.Errorf("Error unmarshaling actual YAML: %v", err)
	}

	if fmt.Sprintf("%v", actualData) != fmt.Sprintf("%v", expectedData) {
		t.Errorf("YAML content does not match expected content")
	}
}

func Test_DeleteNoRecursive(t *testing.T) {
	cmdDeleteNoRecursive := DeleteCmd
	cmdDeleteNoRecursive.SetArgs([]string{
		fmt.Sprintf("-s=%s", "10.10.0.0/20"),
		fmt.Sprintf("-f=%s", "testDeleteNoRecursive.yaml"),
		"-r=false",
	})

	testFile, err := testutils.CreateTestFile("testDeleteNoRecursive.yaml")
	if err != nil {
		t.Errorf("Error creating test file YAML: %v", err)
	}

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error(err)
		}
	}()

	err = cmdDeleteNoRecursive.Execute()
	assert.Equal(t, "cannot delete 10.10.0.0/20 as subnets are defined under it. Use '-r' or '--recursive' to delete 10.10.0.0/20 and everything defined under it", err.Error())
}
