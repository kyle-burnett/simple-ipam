package subnetutils

import (
	"fmt"
	"net"
)

// Check if the subnet from user input is valid
func CheckValidSubnet(subnetToAdd string) error {
	ip, existingNet, err := net.ParseCIDR(subnetToAdd)
	if err != nil {
		return fmt.Errorf("error parsing existing CIDR: %v", err)
	}
	if subnetToAdd != existingNet.String() {
		return fmt.Errorf("%v is not valid CIDR notation", subnetToAdd)
	}
	if ip.To4() == nil {
		return fmt.Errorf("%v is not a valid IPv4 subnet", subnetToAdd)
	}
	return nil
}

// Check if subnetToAdd is a subnet of an existing network
func IsSubnetOf(subnet, subnetToAdd string) (bool, error) {
	_, existingNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return false, fmt.Errorf("error parsing existing subnet: %v", err)
	}

	_, subnetNet, err := net.ParseCIDR(subnetToAdd)
	if err != nil {
		return false, fmt.Errorf("error parsing subnet to add: %v", err)
	}

	existingOnes, _ := existingNet.Mask.Size()
	subnetToAddOnes, _ := subnetNet.Mask.Size()

	if existingNet.Contains(subnetNet.IP) {
		return subnetToAddOnes >= existingOnes, nil
	}

	return false, nil
}

// Check if subnetToAdd is a supernet of existingsubnet
func IsSupernetOf(existingsubnet, subnetToAdd string) (bool, error) {
	_, existingNet, err := net.ParseCIDR(existingsubnet)
	if err != nil {
		return false, fmt.Errorf("error parsing existing subnet: %v", err)
	}

	_, subnetNet, err := net.ParseCIDR(subnetToAdd)
	if err != nil {
		return false, fmt.Errorf("error parsing subnet to add: %v", err)
	}

	existingOnes, _ := existingNet.Mask.Size()
	subnetToAddOnes, _ := subnetNet.Mask.Size()

	if subnetNet.Contains(existingNet.IP) {
		return subnetToAddOnes <= existingOnes, nil
	}

	return false, nil
}
