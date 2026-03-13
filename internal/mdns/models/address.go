package models

import (
	"fmt"
	"net"
)

type ResolvedAddresses struct {
	Name string
	Ip   []net.IP
	Port int
}

func (ra *ResolvedAddresses) GetAddresses() []string {
	var addresses []string
	for _, ip := range ra.Ip {
		addresses = append(addresses, fmt.Sprintf("%s:%d", ip.String(), ra.Port))
	}
	return addresses
}
