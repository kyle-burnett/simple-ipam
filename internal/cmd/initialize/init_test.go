package initialize

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/kyle-burnett/simple-ipam/internal/models"
)

func Test_InitCommand(t *testing.T) {
	err := Initialize("test", "test")
	require.NoError(t, err)
	defer os.Remove("test.yaml")

	ipamFile, err := os.ReadFile("test.yaml")
	require.NoError(t, err)

	expectedYAML := models.IPAM{
		Description: "test",
		Subnets:     map[string]models.Subnets{},
	}
	expectedYamlData, err := yaml.Marshal(&expectedYAML)
	require.NoError(t, err)

	assert.Equal(t, string(expectedYamlData), string(ipamFile))
}

func Test_InitCommand_FileAlreadyExists(t *testing.T) {
	err := Initialize("test", "test")
	require.NoError(t, err)
	defer os.Remove("test.yaml")

	err = Initialize("test", "test")
	assert.EqualError(t, err, "IPAM file test.yaml already exists")
}
