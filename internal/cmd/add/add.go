package add

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/kyle-burnett/simple-ipam/internal/utils/checkvalidcidr"
)

var cidr, description, inputFilename string
var tags []string
var print bool
var ipam IPAM

type IPAM struct {
	IPAM map[string]interface{} `yaml:"ipam"`
}

var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a prefix to an IPAM file",
	Run: func(cmd *cobra.Command, args []string) {
		Add()
	},
}

func init() {
	AddCmd.Flags().StringVarP(&cidr, "cidr", "c", "", "CIDR to Add")
	AddCmd.Flags().StringVarP(&description, "description", "d", "", "CIDR to Add")
	AddCmd.Flags().StringVarP(&inputFilename, "ipam-file", "i", "", "ipam file")
	AddCmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "Tags to add to the CIDR")
	AddCmd.Flags().BoolVarP(&print, "print", "p", false, "Print contents of the IPAM file to stdout")
	AddCmd.MarkFlagRequired("cidr")
	AddCmd.MarkFlagRequired("ipam-file")
}

func Add() {
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
	addCIDR(prefixes, cidr)

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

func checkValidCIDR(cidrToAdd string) {
	_, existingNet, err := net.ParseCIDR(cidrToAdd)
	if err != nil {
		log.Fatalf("Error parsing existing CIDR: %v\n", err)
	}
	if cidr != existingNet.String() {
		log.Fatalf("%v is not valid CIDR notation", cidr)
	}
}

// Add a new CIDR to an IPAM file. This will iterate through the IPAM
// hierarchy to check if cidrToAdd is a subnet of any existing CIDR
// in the IPAM file. If it is, cidrToAdd is added directly under
// the longest CIDR that it is a subnet of.
func addCIDR(prefixes map[string]interface{}, cidrToAdd string) {
	for existingCIDR, existingValue := range prefixes {
		if existingCIDR == cidrToAdd {
			log.Fatalf("%v already exists in this IPAM file.\n", cidrToAdd)
		}
		if isSubnetOf(existingCIDR, cidrToAdd) {
			if subnets, ok := existingValue.(map[string]interface{}); ok {
				subnet_map := subnets["subnets"].(map[string]interface{})
				// If we reach the last value in a given prefix hierarchy (subnets: {}), add cidrToAdd
				// to the subnets key of that prefix
				if len(subnet_map) == 0 {
					subnets["subnets"].(map[string]interface{})[cidrToAdd] = map[string]interface{}{
						"description": description,
						"cidr_tags":   tags,
						"subnets":     make(map[string]interface{}),
					}
					return
				} else {
					addCIDR(subnet_map, cidrToAdd)
				}
				return
			}
		}
	}
	prefixes[cidrToAdd] = map[string]interface{}{
		"description": description,
		"cidr_tags":   tags,
		"subnets":     make(map[string]interface{}),
	}
	// If needed, rearrange the IPAM file to maintain a correct CIDR hierarchy
	// For example, maybe a shorter prefix was added and any longer prefixes need
	// to be moved under it
	rearrangeIPAM(prefixes, cidrToAdd)
}

// Check if cidrToAdd is a subnet of existingCIDR
func isSubnetOf(existingCIDR, cidrToAdd string) (bool bool) {
	_, existingNet, err := net.ParseCIDR(existingCIDR)
	if err != nil {
		log.Fatalf("Error parsing existing CIDR: %v", err)
	}

	_, cidrNet, err := net.ParseCIDR(cidrToAdd)
	if err != nil {
		log.Fatalf("Error parsing CIDR to add: %v", err)
	}

	existingCIDRMask, _ := strconv.Atoi(strings.Split(existingCIDR, "/")[1])
	cidrToAddMask, _ := strconv.Atoi(strings.Split(cidrToAdd, "/")[1])

	if existingNet.Contains(cidrNet.IP) {
		return cidrToAddMask >= existingCIDRMask
	}

	return false
}

// Check if cidrToAdd is a supernet of existingCIDR
func isSupernetOf(existingCIDR, cidrToAdd string) (bool bool) {
	// Parse the existing CIDR
	_, existingNet, err := net.ParseCIDR(existingCIDR)
	if err != nil {
		log.Fatalf("Error parsing existing CIDR: %v", err)
	}

	// Parse the CIDR to add
	_, cidrNet, err := net.ParseCIDR(cidrToAdd)
	if err != nil {
		log.Fatalf("Error parsing CIDR to add: %v", err)
	}

	existingCIDRMask, _ := strconv.Atoi(strings.Split(existingCIDR, "/")[1])
	cidrToAddMask, _ := strconv.Atoi(strings.Split(cidrToAdd, "/")[1])

	if cidrNet.Contains(existingNet.IP) {
		return cidrToAddMask <= existingCIDRMask
	}

	return false
}

// If needed, rearrange the IPAM hierarchy after adding a new CIDR.
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
func rearrangeIPAM(prefixes map[string]interface{}, cidrToAdd string) {
	for existingCIDR := range prefixes {
		// Don't add cidrToAdd under itself
		if existingCIDR == cidrToAdd {
			continue
		}
		if isSupernetOf(existingCIDR, cidrToAdd) {
			childMap, _ := prefixes[existingCIDR].(map[string]interface{})
			if subnets, ok := prefixes[cidrToAdd].(map[string]interface{}); ok {
				subnet_map := subnets["subnets"].(map[string]interface{})
				subnet_map[existingCIDR] = childMap
				delete(prefixes, existingCIDR)
				return
			}
		}
	}
}
