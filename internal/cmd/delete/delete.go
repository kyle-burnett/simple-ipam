package delete

import (
	"log"
	"os"
	"path/filepath"

	"github.com/kyle-burnett/simple-ipam/internal/models"
	"github.com/kyle-burnett/simple-ipam/internal/utils/checkvalidsubnet"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var subnet, inputFile string
var recursive bool
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
	_ = DeleteCmd.MarkFlagRequired("subnet")
	_ = DeleteCmd.MarkFlagRequired("file")
	DeleteCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Delete a CIDR and all subnets under it")
}

func Delete() {
	cleanup := true
	ipamFile, err := os.ReadFile(inputFile)
	if err != nil {
		log.Print("Error reading YAML file:", err)
		return
	}

	err = yaml.Unmarshal(ipamFile, &ipam)
	if err != nil {
		log.Printf("Error unmarshaling YAML: %v", err)
		return
	}

	checkvalidsubnet.CheckValidSubnet(subnet)
	deleteCIDR(ipam.Subnets, subnet)

	updatedYAML, err := yaml.Marshal(&ipam)
	if err != nil {
		log.Printf("Error marshaling YAML: %v", err)
		return
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(inputFile), "tmp_ipam.*.txt")
	if err != nil {
		log.Printf("Error creating temp file: %v", err)
		return
	}
	defer func() {
		if cleanup {
			err := os.Remove(tmpFile.Name())
			if err != nil {
				log.Printf("Error removing temp file: %v", err)
			}
		}
	}()

	_, err = tmpFile.Write(updatedYAML)
	if err != nil {
		tmpFile.Close()
		log.Printf("Error writing to temp file %v:", err)
		return
	}

	if err := tmpFile.Close(); err != nil {
		log.Printf("Error closing temp file %v:", err)
		return
	}

	err = os.Rename(tmpFile.Name(), inputFile)
	if err != nil {
		log.Printf("Error writing IPAM data %v:", err)
		return
	}
	cleanup = false
}

func deleteCIDR(allSubnets map[string]models.Subnets, subnetToDelete string) {
	if _, ok := allSubnets[subnetToDelete]; ok {
		if len(allSubnets[subnetToDelete].Subnets) > 0 && !recursive {
			log.Printf("Cannot delete %[1]s as subnets are defined under it. Use '-r' or '--recursive' to delete %[1]s and everything defined under it", subnetToDelete)
		} else {
			delete(allSubnets, subnetToDelete)
		}
	}
	for _, v := range allSubnets {
		if _, ok := v.Subnets[subnetToDelete]; ok {
			if len((v.Subnets)[subnetToDelete].Subnets) > 0 && !recursive {
				log.Printf("Cannot delete %[1]s as subnets are defined under it. Use '-r' or '--recursive' to delete %[1]s and everything defined under it", subnetToDelete)
			} else {
				delete(v.Subnets, subnetToDelete)
			}
		} else {
			deleteCIDR(v.Subnets, subnetToDelete)
		}
	}
}
