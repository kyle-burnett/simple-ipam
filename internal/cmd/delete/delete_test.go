package delete

import (
	"os"
	"testing"

	"github.com/kyle-burnett/simple-ipam/internal/models"
	"github.com/kyle-burnett/simple-ipam/internal/utils/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func Test_Delete(t *testing.T) {
	testFile, err := testutils.CreateTestFile("testDelete.yaml")
	require.NoError(t, err)
	defer os.Remove(testFile)

	err = Delete(testFile, "10.10.0.0/24", false)
	require.NoError(t, err)

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
	require.NoError(t, err)

	ipamFile, err := os.ReadFile(testFile)
	require.NoError(t, err)

	assert.Equal(t, string(expectedYamlData), string(ipamFile))
}

func Test_DeleteRecursive(t *testing.T) {
	testFile, err := testutils.CreateTestFile("testDeleteRecursive.yaml")
	require.NoError(t, err)
	defer os.Remove(testFile)

	err = Delete(testFile, "10.10.0.0/20", true)
	require.NoError(t, err)

	expectedYAML := models.IPAM{
		Subnets: map[string]models.Subnets{},
	}

	expectedYamlData, err := yaml.Marshal(&expectedYAML)
	require.NoError(t, err)

	ipamFile, err := os.ReadFile(testFile)
	require.NoError(t, err)

	assert.Equal(t, string(expectedYamlData), string(ipamFile))
}

func Test_DeleteNoRecursive(t *testing.T) {
	testFile, err := testutils.CreateTestFile("testDeleteNoRecursive.yaml")
	require.NoError(t, err)
	defer os.Remove(testFile)

	err = Delete(testFile, "10.10.0.0/20", false)
	assert.EqualError(t, err, "cannot delete 10.10.0.0/20 as subnets are defined under it. Use '-r' or '--recursive' to delete 10.10.0.0/20 and everything defined under it")
}
