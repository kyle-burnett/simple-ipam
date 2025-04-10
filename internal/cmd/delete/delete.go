package delete

import (
	"fmt"
	"log"
	"os"

	"github.com/kyle-burnett/simple-ipam/internal/models"
	"github.com/kyle-burnett/simple-ipam/internal/utils/checkvalidsubnet"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var subnet, inputFile string
var print, force bool
var ipam models.IPAM

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a prefix from an IPAM file",
	Run: func(cmd *cobra.Command, args []string) {
		Delete()
	},
}

func init() {
	DeleteCmd.Flags().StringVarP(&subnet, "subnet", "s", "", "subnet to Add")
	DeleteCmd.Flags().StringVarP(&inputFile, "file", "f", "", "ipam file")
	DeleteCmd.MarkFlagRequired("subnet")
	DeleteCmd.MarkFlagRequired("file")
	DeleteCmd.Flags().BoolVarP(&print, "print", "p", false, "Print contents of the IPAM file to stdout")
	DeleteCmd.Flags().BoolVarP(&force, "recursive", "r", false, "Delete a CIDR and all subnets under it")
	DeleteCmd.MarkFlagRequired("subnet")
	DeleteCmd.MarkFlagRequired("ipam-file")
}

func Delete() {
	ipamFile, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatal("Error reading YAML file:", err)
	}

	err = yaml.Unmarshal(ipamFile, &ipam)
	if err != nil {
		log.Fatalf("Error unmarshaling YAML: %v", err)
	}

	checkvalidsubnet.CheckValidSubnet(subnet)
	deleteCIDR(ipam.Subnets, subnet)

	updatedYAML, err := yaml.Marshal(&ipam)
	if err != nil {
		log.Fatalf("Error marshaling YAML: %v", err)
	}

	if print {
		fmt.Println(string(updatedYAML))
	}

	err = os.WriteFile(inputFile, updatedYAML, 0644)
	if err != nil {
		log.Fatalf("Error writing YAML file: %v", err)
	}
}

func deleteCIDR(allSubnets map[string]models.Subnets, subnetToDelete string) {
	if _, ok := allSubnets[subnetToDelete]; ok {
		if len(allSubnets[subnetToDelete].Subnets) > 0 && !force {
			log.Fatalf("Cannot delete %[1]s as subnets are defined under it. Use '-r' or '--recursive' to delete %[1]s and everything defined under it", subnetToDelete)
		} else {
			delete(allSubnets, subnetToDelete)
		}
	}
	for _, v := range allSubnets {
		if _, ok := v.Subnets[subnetToDelete]; ok {
			if len((v.Subnets)[subnetToDelete].Subnets) > 0 && !force {
				log.Fatalf("Cannot delete %[1]s as subnets are defined under it. Use '-r' or '--recursive' to delete %[1]s and everything defined under it", subnetToDelete)
			} else {
				delete(v.Subnets, subnetToDelete)
			}
		} else {
			deleteCIDR(v.Subnets, subnetToDelete)
		}
	}
}

// func checkForSubnets(prefixes interface{}) bool {
// 	if subnets, ok := prefixes.(map[string]interface{}); ok {
// 		subnet_map := subnets["subnets"].(map[string]interface{})
// 		return len(subnet_map) != 0
// 	}
// 	return true
// }
