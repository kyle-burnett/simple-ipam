package delete

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kyle-burnett/simple-ipam/internal/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var subnet, inputFile string
var recursive bool
var ipam models.IPAM

var DeleteCmd = &cobra.Command{
	Use:          "delete",
	Short:        "Delete a prefix from an IPAM file",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := Delete()
		return err
	},
}

func init() {
	DeleteCmd.Flags().StringVarP(&subnet, "subnet", "s", "", "subnet to Add")
	DeleteCmd.Flags().StringVarP(&inputFile, "file", "f", "", "ipam file")
	_ = DeleteCmd.MarkFlagRequired("subnet")
	_ = DeleteCmd.MarkFlagRequired("file")
	DeleteCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Delete a CIDR and all subnets under it")
}

func Delete() error {
	cleanup := true
	ipamFile, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("error reading YAML file: %v", err)
	}

	err = yaml.Unmarshal(ipamFile, &ipam)
	if err != nil {
		return fmt.Errorf("error unmarshaling YAML: %v", err)
	}

	err = deleteCIDR(ipam.Subnets, subnet)
	if err != nil {
		return err
	}

	updatedYAML, err := yaml.Marshal(&ipam)
	if err != nil {
		return fmt.Errorf("error marshaling YAML: %v", err)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(inputFile), "tmp_ipam.*.txt")
	if err != nil {
		return fmt.Errorf("error creating temp file: %v", err)
	}
	defer func() {
		if cleanup {
			err := os.Remove(tmpFile.Name())
			if err != nil {
				log.Printf("error removing temp file: %v", err)
			}
		}
	}()

	_, err = tmpFile.Write(updatedYAML)
	if err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("error writing to temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("error closing temp file: %v", err)
	}

	err = os.Rename(tmpFile.Name(), inputFile)
	if err != nil {
		return fmt.Errorf("error writing IPAM data: %v", err)
	}
	cleanup = false
	return nil
}

func deleteCIDR(allSubnets map[string]models.Subnets, subnetToDelete string) error {
	if _, ok := allSubnets[subnetToDelete]; ok {
		if len(allSubnets[subnetToDelete].Subnets) > 0 && !recursive {
			return fmt.Errorf("cannot delete %[1]s as subnets are defined under it. Use '-r' or '--recursive' to delete %[1]s and everything defined under it", subnetToDelete)
		} else {
			delete(allSubnets, subnetToDelete)
		}
	}
	for _, v := range allSubnets {
		if _, ok := v.Subnets[subnetToDelete]; ok {
			if len((v.Subnets)[subnetToDelete].Subnets) > 0 && !recursive {
				return fmt.Errorf("cannot delete %[1]s as subnets are defined under it. Use '-r' or '--recursive' to delete %[1]s and everything defined under it", subnetToDelete)
			} else {
				delete(v.Subnets, subnetToDelete)
			}
		} else {
			err := deleteCIDR(v.Subnets, subnetToDelete)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
