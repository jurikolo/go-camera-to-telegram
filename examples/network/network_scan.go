// Package main demonstrates how to use the network scanning functionality
package main

import (
	"fmt"
	"log"

	"github.com/jurikolo/go-camera-to-telegram/pkg/network"
)

func main() {
	// Example of parsing a CIDR range
	cidr := "192.168.1.0/24"
	ips, err := network.ParseCIDR(cidr)
	if err != nil {
		log.Fatalf("Failed to parse CIDR: %v", err)
	}

	fmt.Printf("IPs in range %s:\n", cidr)
	for _, ip := range ips {
		fmt.Printf("  %s\n", ip)
	}

	// Example of checking if a port is open
	host := "8.8.8.8"
	port := "53" // DNS
	if network.IsPortOpen(host, port, 5) {
		fmt.Printf("Port %s is open on %s\n", port, host)
	} else {
		fmt.Printf("Port %s is closed on %s\n", port, host)
	}

	// Example of getting local IP
	localIP, err := network.GetLocalIP()
	if err != nil {
		log.Printf("Failed to get local IP: %v", err)
	} else {
		fmt.Printf("Local IP: %s\n", localIP)
	}
}
