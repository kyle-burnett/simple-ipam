package delete

import (
	"fmt"
	"log"
	"os"

	"github.com/kyle-burnett/simple-ipam/internal/utils/checkvalidcidr"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var cidr, inputFilename string
var print, force bool
var ipam IPAM

type IPAM struct {
	IPAM map[string]interface{} `yaml:"ipam"`
}

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a prefix from an IPAM file",
	Run: func(cmd *cobra.Command, args []string) {
		Delete()
	},
}

func init() {
	DeleteCmd.Flags().StringVarP(&cidr, "cidr", "c", "", "CIDR to delete")
	DeleteCmd.Flags().StringVarP(&inputFilename, "ipam-file", "i", "", "ipam file")
	DeleteCmd.Flags().BoolVarP(&print, "print", "p", false, "Print contents of the IPAM file to stdout")
	DeleteCmd.Flags().BoolVarP(&force, "force", "f", false, "Delete a CIDR and all subnets under it")
	err := DeleteCmd.MarkFlagRequired("cidr")
	if err != nil {
		log.Fatal(err)
	}
	err = DeleteCmd.MarkFlagRequired("ipam-file")
	if err != nil {
		log.Fatal(err)
	}
}

func Delete() {
	ipamFile, err := os.ReadFile(inputFilename)
	if err != nil {
		log.Fatal("Error reading YAML file:", err)
	}

	err = yaml.Unmarshal(ipamFile, &ipam)
	if err != nil {
		log.Fatalf("Error unmarshaling YAML: %v", err)
	}

	prefixes, ok := ipam.IPAM["prefixes"].(map[string]interface{})
	if !ok {
		log.Fatal("interface conversion: interface {} is nil, not map[string]interface {}")
	}

	checkvalidcidr.CheckValidCIDR(cidr)
	deleteCIDR(prefixes, cidr)

	updatedYAML, err := yaml.Marshal(&ipam)
	if err != nil {
		log.Fatalf("Error marshaling YAML: %v", err)
	}

	if print {
		fmt.Println(string(updatedYAML))
	}

	err = os.WriteFile(inputFilename, updatedYAML, 0644)
	if err != nil {
		log.Fatalf("Error writing YAML file: %v", err)
	}
}

func deleteCIDR(prefixes map[string]interface{}, cidrToDelete string) {
	if _, ok := prefixes[cidrToDelete]; ok {
		subnetsExist := checkForSubnets(prefixes[cidrToDelete])
		if subnetsExist && !force {
			log.Fatalf("Cannot delete %v as subnets are defined under it. Use '-f' or '--force' to delete %v and everything defined under it", cidr, cidr)
		} else {
			delete(prefixes, cidrToDelete)
		}
	}
	for _, v := range prefixes {
		if subdata, ok := v.(map[string]interface{}); ok {
			deleteCIDR(subdata, cidrToDelete)
		}
	}
}

func checkForSubnets(prefixes interface{}) bool {
	if subnets, ok := prefixes.(map[string]interface{}); ok {
		subnet_map := subnets["subnets"].(map[string]interface{})
		return len(subnet_map) != 0
	}
	return true
}
