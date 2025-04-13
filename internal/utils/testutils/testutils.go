package testutils

import (
	"log"
	"os"

	"github.com/kyle-burnett/simple-ipam/internal/models"
	"gopkg.in/yaml.v3"
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
		log.Printf("Error while Marshaling. %v", err)
		return "", err
	}

	if _, err = os.Stat(fileName); err == nil {
		log.Print("File already exists")
		return "", err
	}

	err = os.WriteFile(fileName, yamlData, 0644)
	if err != nil {
		log.Print("Unable to write data into the file")
		return "", err
	}

	return fileName, nil
}
