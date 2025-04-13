package initialize

import (
	"errors"
	"io/fs"
	"log"
	"os"

	"github.com/kyle-burnett/simple-ipam/internal/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var file, description string

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize an empty IPAM file",
	Run: func(cmd *cobra.Command, args []string) {
		Initialize()
	},
}

func init() {
	InitCmd.Flags().StringVarP(&file, "file", "f", "ipam", "Root IPAM file to create")
	InitCmd.Flags().StringVarP(&description, "description", "d", "", "Root IPAM file description")
}

func Initialize() {
	ipam := models.IPAM{
		Subnets:     make(map[string]models.Subnets),
		Description: description,
	}

	yamlData, err := yaml.Marshal(&ipam)
	if err != nil {
		log.Printf("Error while marshaling YAML: %v", err)
		return
	}

	fileName := file + ".yaml"
	if _, err = os.Stat(fileName); errors.Is(err, fs.ErrNotExist) {
		err = os.WriteFile(fileName, yamlData, 0644)
		if err != nil {
			log.Printf("Unable to create IPAM file: %v", err)
		}
	} else if err == nil {
		log.Printf("IPAM file %v already exists", fileName)
	} else {
		log.Print(err)
	}
}
