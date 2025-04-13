package add

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/kyle-burnett/simple-ipam/internal/models"
)

var subnet, description, inputFile string
var tags []string
var ipam models.IPAM

var AddCmd = &cobra.Command{
	Use:          "add",
	Short:        "Add a subnet to an IPAM file",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := Add()
		return err
	},
}

func init() {
	AddCmd.Flags().StringVarP(&subnet, "subnet", "s", "", "subnet to Add")
	AddCmd.Flags().StringVarP(&inputFile, "file", "f", "", "ipam file")
	_ = AddCmd.MarkFlagRequired("subnet")
	_ = AddCmd.MarkFlagRequired("file")
	AddCmd.Flags().StringVarP(&description, "description", "d", "", "subnet to Add")
	AddCmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "Tags to add to the subnet")
}

func Add() error {
	cleanup := true
	ipamData, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("error reading IPAM file: %v", err)
	}

	err = yaml.Unmarshal(ipamData, &ipam)
	if err != nil {
		return fmt.Errorf("error unmarshaling IPAM: %v", err)
	}

	err = checkValidSubnet(subnet)
	if err != nil {
		return fmt.Errorf("invalid subnet: %v", err)
	}
	err = addsubnet(ipam.Subnets, subnet)
	if err != nil {
		return fmt.Errorf("error adding subnet: %v", err)
	}

	updatedYAML, err := yaml.Marshal(&ipam)
	if err != nil {
		return fmt.Errorf("error marshaling IPAM: %v", err)
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

// Add a subnet to an IPAM file.
func addsubnet(allSubnets map[string]models.Subnets, subnetToAdd string) error {
	for subnet, values := range allSubnets {
		if subnet == subnetToAdd {
			return fmt.Errorf("%#v already exists in this IPAM file", subnetToAdd)
		}
		isSubnet, err := isSubnetOf(subnet, subnetToAdd)
		if err != nil {
			return err
		}
		if isSubnet {
			// We reached the end. No need to continue checking.
			if len(values.Subnets) == 0 {
				if _, ok := values.Subnets[subnetToAdd]; !ok {
					values.Subnets[subnetToAdd] = models.Subnets{
						Description: description,
						Tags:        tags,
						Subnets:     map[string]models.Subnets{},
					}
				}
				return nil
			} else {
				err := addsubnet(values.Subnets, subnetToAdd)
				if err != nil {
					return err
				}
			}
			return nil
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

// Check if the subnet from user input is valid
func checkValidSubnet(subnetToAdd string) error {
	_, existingNet, err := net.ParseCIDR(subnetToAdd)
	if err != nil {
		return fmt.Errorf("error parsing existing CIDR: %v", err)
	}
	if subnetToAdd != existingNet.String() {
		return fmt.Errorf("%v is not valid CIDR notation", subnetToAdd)
	}
	return nil
}

// Check if subnetToAdd is a subnet of an existing network
func isSubnetOf(subnet, subnetToAdd string) (bool, error) {
	_, existingNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return false, fmt.Errorf("error parsing existing subnet: %v", err)
	}

	_, subnetNet, err := net.ParseCIDR(subnetToAdd)
	if err != nil {
		return false, fmt.Errorf("error parsing subnet to add: %v", err)
	}

	existingsubnetMask, _ := strconv.Atoi(strings.Split(subnet, "/")[1])
	subnetToAddMask, _ := strconv.Atoi(strings.Split(subnetToAdd, "/")[1])

	if existingNet.Contains(subnetNet.IP) {
		return subnetToAddMask >= existingsubnetMask, nil
	}

	return false, nil
}

// Check if subnetToAdd is a supernet of existingsubnet
func isSupernetOf(existingsubnet, subnetToAdd string) (bool, error) {
	// Parse the existing subnet
	_, existingNet, err := net.ParseCIDR(existingsubnet)
	if err != nil {
		return false, fmt.Errorf("error parsing existing subnet: %v", err)
	}

	// Parse the subnet to add
	_, subnetNet, err := net.ParseCIDR(subnetToAdd)
	if err != nil {
		return false, fmt.Errorf("error parsing subnet to add: %v", err)
	}

	existingsubnetMask, _ := strconv.Atoi(strings.Split(existingsubnet, "/")[1])
	subnetToAddMask, _ := strconv.Atoi(strings.Split(subnetToAdd, "/")[1])

	if subnetNet.Contains(existingNet.IP) {
		return subnetToAddMask <= existingsubnetMask, nil
	}

	return false, nil
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
		isSupernet, err := isSupernetOf(subnet, subnetToAdd)
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
