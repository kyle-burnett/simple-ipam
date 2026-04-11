package add

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"

	"github.com/kyle-burnett/simple-ipam/internal/models"
	"github.com/kyle-burnett/simple-ipam/internal/utils/fileutil"
	"github.com/kyle-burnett/simple-ipam/internal/utils/subnetutils"
)

var subnet, description, inputFile string
var tags []string

var AddCmd = &cobra.Command{
	Use:          "add",
	Short:        "Add a subnet to an IPAM file",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return Add(inputFile, subnet, description, tags)
	},
}

func init() {
	AddCmd.Flags().StringVarP(&subnet, "subnet", "s", "", "subnet to Add")
	AddCmd.Flags().StringVarP(&inputFile, "file", "f", "", "ipam file")
	_ = AddCmd.MarkFlagRequired("subnet")
	_ = AddCmd.MarkFlagRequired("file")
	AddCmd.Flags().StringVarP(&description, "description", "d", "", "description for the subnet")
	AddCmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "Tags to add to the subnet")
}

func Add(inputFile, subnet, description string, tags []string) error {
	ipamData, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("error reading IPAM file: %v", err)
	}

	var ipam models.IPAM
	err = yaml.Unmarshal(ipamData, &ipam)
	if err != nil {
		return fmt.Errorf("error unmarshaling IPAM: %v", err)
	}

	err = subnetutils.CheckValidSubnet(subnet)
	if err != nil {
		return fmt.Errorf("invalid subnet: %v", err)
	}
	err = addsubnet(ipam.Subnets, subnet, description, tags)
	if err != nil {
		return fmt.Errorf("error adding subnet: %v", err)
	}

	return fileutil.WriteYAMLAtomic(inputFile, &ipam)
}

// Add a subnet to an IPAM file.
func addsubnet(allSubnets map[string]models.Subnets, subnetToAdd, description string, tags []string) error {
	for subnet, values := range allSubnets {
		if subnet == subnetToAdd {
			return fmt.Errorf("%#v already exists in this IPAM file", subnetToAdd)
		}
		isSubnet, err := subnetutils.IsSubnetOf(subnet, subnetToAdd)
		if err != nil {
			return err
		}
		if isSubnet {
			if len(values.Subnets) == 0 {
				values.Subnets[subnetToAdd] = models.Subnets{
					Description: description,
					Tags:        tags,
					Subnets:     map[string]models.Subnets{},
				}
				return nil
			}
			return addsubnet(values.Subnets, subnetToAdd, description, tags)
		}
	}
	allSubnets[subnetToAdd] = models.Subnets{
		Description: description,
		Tags:        tags,
		Subnets:     map[string]models.Subnets{},
	}
	// Re-arrange the IPAM file to keep the newly added subnet in order
	err := rearrangeSubnets(allSubnets, subnetToAdd)
	if err != nil {
		return err
	}
	return nil
}

// Re-arrange the IPAM hierarchy after adding a new subnet.
// For example if we have:
//
//	prefixes:
//		10.10.0.0/20:
//			10.10.0.0/22:
//				10.10.0.0/24:
//
// and we add '10.10.0.0/21', we should end up with:
//
//	prefixes:
//		10.10.0.0/20:
//			10.10.0.0/21:
//				10.10.0.0/22:
//					10.10.0.0/24:
func rearrangeSubnets(allSubnets map[string]models.Subnets, subnetToAdd string) error {
	for subnet, values := range allSubnets {
		// Don't add subnetToAdd under itself
		if subnet == subnetToAdd {
			continue
		}
		isSupernet, err := subnetutils.IsSupernetOf(subnet, subnetToAdd)
		if err != nil {
			return err
		}
		if isSupernet {
			childMap := values
			if subnets, ok := allSubnets[subnetToAdd]; ok {
				subnetMap := subnets.Subnets
				subnetMap[subnet] = childMap
				delete(allSubnets, subnet)
				return nil
			}
		}
	}
	return nil
}
