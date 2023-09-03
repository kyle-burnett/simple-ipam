package initialize

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var file, description string

type IPAM struct {
	IPAM map[string]interface{}
}

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize an empty IPAM file",
	Run: func(cmd *cobra.Command, args []string) {
		Initialize()
	},
}

func init() {
	InitCmd.Flags().StringVarP(&file, "file", "f", "", "Root IPAM file to create")
	InitCmd.Flags().StringVarP(&description, "description", "d", "", "Root IPAM file description")
	InitCmd.MarkFlagRequired("file")
	InitCmd.MarkFlagRequired("description")
}

func Initialize() {
	ipam := IPAM{
		IPAM: make(map[string]interface{}),
	}

	ipam.IPAM["description"] = description

	prefixes := make(map[string]interface{})
	ipam.IPAM["prefixes"] = prefixes

	yamlData, err := yaml.Marshal(&ipam)

	if err != nil {
		fmt.Printf("Error while Marshaling. %v", err)
	}

	fileName := file + ".yaml"

	if _, err = os.Stat(fileName); err == nil {
		log.Fatal("File already exists")
	}

	err = os.WriteFile(fileName, yamlData, 0644)
	if err != nil {
		panic("Unable to write data into the file")
	}
}
