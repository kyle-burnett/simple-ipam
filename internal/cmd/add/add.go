package add

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"gopkg.in/yaml.v3"

	"github.com/kyle-burnett/simple-ipam/internal/models"
	"github.com/kyle-burnett/simple-ipam/internal/utils/checkvalidsubnet"
)

var subnet, description, inputFile string
var tags []string
var print bool
var ipam models.IPAM

var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a subnet to an IPAM file",
	Run: func(cmd *cobra.Command, args []string) {
		err := doc.GenMarkdownTree(cmd, "./docs")
		if err != nil {
			log.Fatal(err)
		}
		Add()
	},
}

func init() {
	AddCmd.Flags().StringVarP(&subnet, "subnet", "s", "", "subnet to Add")
	AddCmd.Flags().StringVarP(&inputFile, "file", "f", "", "ipam file")
	AddCmd.MarkFlagRequired("subnet")
	AddCmd.MarkFlagRequired("file")
	AddCmd.Flags().StringVarP(&description, "description", "d", "", "subnet to Add")
	AddCmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "Tags to add to the subnet")
	AddCmd.Flags().BoolVarP(&print, "print", "p", false, "Print contents of the IPAM file to stdout")
}

func Add() {
	ipamFile, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("Error reading YAML file %v:", err)
	}

	err = yaml.Unmarshal(ipamFile, &ipam)
	if err != nil {
		log.Fatalf("Error unmarshaling YAML: %v", err)
	}

	checkvalidsubnet.CheckValidSubnet(subnet)
	allSubnets := ipam.Subnets
	addsubnet(allSubnets, subnet)

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

// Add a new subnet to an IPAM file.
func addsubnet(allSubnets map[string]models.Subnets, subnetToAdd string) {
	for subnet, values := range allSubnets {
		if subnet == subnetToAdd {
			log.Fatalf("%#v already exists in this IPAM file.\n", subnetToAdd)
		}
		if isSubnetOf(subnet, subnetToAdd) {
			// We reached the end. No need to continue checking.
			if len(values.Subnets) == 0 {
				if _, ok := values.Subnets[subnetToAdd]; !ok {
					values.Subnets[subnetToAdd] = models.Subnets{
						Description: description,
						Tags:        tags,
						Subnets:     map[string]models.Subnets{},
					}
				}
				return
			} else {
				addsubnet(values.Subnets, subnetToAdd)
			}
			return
		}
	}
	allSubnets[subnetToAdd] = models.Subnets{
		Description: description,
		Tags:        tags,
		Subnets:     map[string]models.Subnets{},
	}
	// Re-arrange the IPAM file to keep the newly added subnet in order
	rearrangeSubnets(allSubnets, subnetToAdd)
}

// Check if subnetToAdd is a subnet of parent subnet
func isSubnetOf(subnet, subnetToAdd string) bool {
	_, existingNet, err := net.ParseCIDR(subnet)
	if err != nil {
		log.Fatalf("Error parsing existing subnet: %v", err)
	}

	_, subnetNet, err := net.ParseCIDR(subnetToAdd)
	if err != nil {
		log.Fatalf("Error parsing subnet to add: %v", err)
	}

	existingsubnetMask, _ := strconv.Atoi(strings.Split(subnet, "/")[1])
	subnetToAddMask, _ := strconv.Atoi(strings.Split(subnetToAdd, "/")[1])

	if existingNet.Contains(subnetNet.IP) {
		return subnetToAddMask >= existingsubnetMask
	}

	return false
}

// Check if subnetToAdd is a supernet of existingsubnet
func isSupernetOf(existingsubnet, subnetToAdd string) (bool bool) {
	// Parse the existing subnet
	_, existingNet, err := net.ParseCIDR(existingsubnet)
	if err != nil {
		log.Fatalf("Error parsing existing subnet: %v", err)
	}

	// Parse the subnet to add
	_, subnetNet, err := net.ParseCIDR(subnetToAdd)
	if err != nil {
		log.Fatalf("Error parsing subnet to add: %v", err)
	}

	existingsubnetMask, _ := strconv.Atoi(strings.Split(existingsubnet, "/")[1])
	subnetToAddMask, _ := strconv.Atoi(strings.Split(subnetToAdd, "/")[1])

	if subnetNet.Contains(existingNet.IP) {
		return subnetToAddMask <= existingsubnetMask
	}

	return false
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
func rearrangeSubnets(allSubnets map[string]models.Subnets, subnetToAdd string) {
	for subnet, values := range allSubnets {
		// Don't add subnetToAdd under itself
		if subnet == subnetToAdd {
			continue
		}
		if isSupernetOf(subnet, subnetToAdd) {
			childMap := values
			if subnets, ok := allSubnets[subnetToAdd]; ok {
				subnet_map := subnets.Subnets
				subnet_map[subnet] = childMap
				delete(allSubnets, subnet)
				return
			}
		}
	}
}
