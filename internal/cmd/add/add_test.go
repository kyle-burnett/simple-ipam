package add

import (
	"fmt"
	"os"
	"testing"

	"github.com/kyle-burnett/simple-ipam/internal/models"
	"github.com/kyle-burnett/simple-ipam/internal/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func Test_AddSubnet(t *testing.T) {
	testFile, err := testutils.CreateTestFile("testAdd.yaml")
	require.NoError(t, err)
	defer os.Remove(testFile)

	err = Add(testFile, "10.10.0.0/25", "test subnet", []string{})
	require.NoError(t, err)

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
	require.NoError(t, err)

	ipamFile, err := os.ReadFile(testFile)
	require.NoError(t, err)

	assert.Equal(t, string(expectedYamlData), string(ipamFile))
}

func Test_AddSupernet(t *testing.T) {
	testFile, err := testutils.CreateTestFile("testSupernet.yaml")
	require.NoError(t, err)
	defer os.Remove(testFile)

	err = Add(testFile, "10.10.0.0/22", "test subnet", []string{})
	require.NoError(t, err)

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
	require.NoError(t, err)

	ipamFile, err := os.ReadFile(testFile)
	require.NoError(t, err)

	assert.Equal(t, string(expectedYamlData), string(ipamFile))
}

func Test_InvalidSubnet(t *testing.T) {
	testFile, err := testutils.CreateTestFile("testInvalidSubnet.yaml")
	require.NoError(t, err)
	defer os.Remove(testFile)

	err = Add(testFile, "10.10.0.0/222", "test subnet", []string{})
	assert.EqualError(t, err, "invalid subnet: error parsing existing CIDR: invalid CIDR address: 10.10.0.0/222")
}

func Test_InvalidNotation(t *testing.T) {
	testFile, err := testutils.CreateTestFile("testInvalidNotation.yaml")
	require.NoError(t, err)
	defer os.Remove(testFile)

	err = Add(testFile, "10.10.0.100/22", "test subnet", []string{})
	assert.EqualError(t, err, fmt.Sprintf("invalid subnet: %s is not valid CIDR notation", "10.10.0.100/22"))
}

func Test_DuplicateSubnet(t *testing.T) {
	testFile, err := testutils.CreateTestFile("testDuplicateSubnet.yaml")
	require.NoError(t, err)
	defer os.Remove(testFile)

	err = Add(testFile, "10.10.0.0/20", "test subnet", []string{})
	assert.EqualError(t, err, "error adding subnet: \"10.10.0.0/20\" already exists in this IPAM file")
}
