package delete

import (
	"fmt"
	"os"

	"github.com/kyle-burnett/simple-ipam/internal/models"
	"github.com/kyle-burnett/simple-ipam/internal/utils/fileutil"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var subnet, inputFile string
var recursive bool

var DeleteCmd = &cobra.Command{
	Use:          "delete",
	Short:        "Delete a prefix from an IPAM file",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return Delete(inputFile, subnet, recursive)
	},
}

func init() {
	DeleteCmd.Flags().StringVarP(&subnet, "subnet", "s", "", "subnet to Add")
	DeleteCmd.Flags().StringVarP(&inputFile, "file", "f", "", "ipam file")
	_ = DeleteCmd.MarkFlagRequired("subnet")
	_ = DeleteCmd.MarkFlagRequired("file")
	DeleteCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Delete a CIDR and all subnets under it")
}

func Delete(inputFile, subnet string, recursive bool) error {
	ipamFile, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("error reading YAML file: %v", err)
	}

	var ipam models.IPAM
	err = yaml.Unmarshal(ipamFile, &ipam)
	if err != nil {
		return fmt.Errorf("error unmarshaling YAML: %v", err)
	}

	err = deleteCIDR(ipam.Subnets, subnet, recursive)
	if err != nil {
		return err
	}

	return fileutil.WriteYAMLAtomic(inputFile, &ipam)
}

func deleteCIDR(allSubnets map[string]models.Subnets, subnetToDelete string, recursive bool) error {
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
			err := deleteCIDR(v.Subnets, subnetToDelete, recursive)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
