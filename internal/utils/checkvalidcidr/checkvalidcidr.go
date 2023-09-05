package checkvalidcidr

import (
	"log"
	"net"
)

func CheckValidCIDR(cidrToAdd string) {
	_, existingNet, err := net.ParseCIDR(cidrToAdd)
	if err != nil {
		log.Fatalf("Error parsing existing CIDR: %v\n", err)
	}
	if cidrToAdd != existingNet.String() {
		log.Fatalf("%v is not valid CIDR notation", cidrToAdd)
	}
}
