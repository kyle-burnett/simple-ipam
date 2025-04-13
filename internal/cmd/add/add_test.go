package add

import (
	"fmt"
	"os"
	"testing"

	"github.com/kyle-burnett/simple-ipam/internal/models"
	"github.com/kyle-burnett/simple-ipam/internal/utils/testutils"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func Test_AddSubnet(t *testing.T) {
	cmdSubnet := AddCmd
	cmdSubnet.SetArgs([]string{
		fmt.Sprintf("-d=%s", "test subnet"),
		fmt.Sprintf("-s=%s", "10.10.0.0/25"),
		fmt.Sprintf("-f=%s", "testAdd.yaml"),
	})

	testFile, err := testutils.CreateTestFile("testAdd.yaml")
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

func Test_AddSupernet(t *testing.T) {
	cmdSupernet := AddCmd
	cmdSupernet.SetArgs([]string{
		fmt.Sprintf("-d=%s", "test subnet"),
		fmt.Sprintf("-s=%s", "10.10.0.0/22"),
		fmt.Sprintf("-f=%s", "testSupernet.yaml"),
	})

	testFile, err := testutils.CreateTestFile("testSupernet.yaml")
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

func Test_InvalidSubnet(t *testing.T) {
	cmdInvalidSubnet := AddCmd
	cmdInvalidSubnet.SetArgs([]string{
		fmt.Sprintf("-d=%s", "test subnet"),
		fmt.Sprintf("-s=%s", "10.10.0.0/222"),
		fmt.Sprintf("-f=%s", "testInvalidSubnet.yaml"),
	})

	testFile, err := testutils.CreateTestFile("testInvalidSubnet.yaml")
	if err != nil {
		t.Errorf("error creating test file YAML: %v", err)
	}

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error(err)
		}
	}()

	err = cmdInvalidSubnet.Execute()
	assert.Equal(t, "invalid subnet: error parsing existing CIDR: invalid CIDR address: 10.10.0.0/222", err.Error())
}

func Test_InvalidNotation(t *testing.T) {
	cmdInvalidSubnet := AddCmd
	cmdInvalidSubnet.SetArgs([]string{
		fmt.Sprintf("-d=%s", "test subnet"),
		fmt.Sprintf("-s=%s", "10.10.0.100/22"),
		fmt.Sprintf("-f=%s", "testInvaliNotation.yaml"),
	})

	testFile, err := testutils.CreateTestFile("testInvaliNotation.yaml")
	if err != nil {
		t.Errorf("error creating test file YAML: %v", err)
	}

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error(err)
		}
	}()

	err = cmdInvalidSubnet.Execute()
	assert.Equal(t, "invalid subnet: 10.10.0.100/22 is not valid CIDR notation", err.Error())
}

func Test_DuplicateSubnet(t *testing.T) {
	cmdInvalidSubnet := AddCmd
	cmdInvalidSubnet.SetArgs([]string{
		fmt.Sprintf("-d=%s", "test subnet"),
		fmt.Sprintf("-s=%s", "10.10.0.0/20"),
		fmt.Sprintf("-f=%s", "testDuplicateSubnet.yaml"),
	})

	testFile, err := testutils.CreateTestFile("testDuplicateSubnet.yaml")
	if err != nil {
		t.Errorf("error creating test file YAML: %v", err)
	}

	defer func() {
		err := os.Remove(testFile)
		if err != nil {
			t.Error(err)
		}
	}()

	err = cmdInvalidSubnet.Execute()
	assert.Equal(t, "error adding subnet: \"10.10.0.0/20\" already exists in this IPAM file", err.Error())
}
