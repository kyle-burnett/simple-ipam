package testutils

import (
	"fmt"
	"os"

	"github.com/kyle-burnett/simple-ipam/internal/models"
	"go.yaml.in/yaml/v4"
)

func CreateTestFile(fileName string) (string, error) {
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
		return "", fmt.Errorf("error marshaling YAML: %v", err)
	}

	if _, err = os.Stat(fileName); err == nil {
		return "", fmt.Errorf("file already exists: %s", fileName)
	}

	err = os.WriteFile(fileName, yamlData, 0644)
	if err != nil {
		return "", fmt.Errorf("error writing test file: %v", err)
	}

	return fileName, nil
}
