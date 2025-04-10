package checkvalidsubnet

import (
	"log"
	"net"
)

func CheckValidSubnet(subnetToAdd string) {
	_, existingNet, err := net.ParseCIDR(subnetToAdd)
	if err != nil {
		log.Fatalf("Error parsing existing CIDR: %v\n", err)
	}
	if subnetToAdd != existingNet.String() {
		log.Fatalf("%v is not valid CIDR notation", subnetToAdd)
	}
}
