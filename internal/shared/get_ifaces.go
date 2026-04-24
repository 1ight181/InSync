package resolver

import (
	"net"
	"strconv"
)

func GetNetworkInterfaces(interfaces []string) ([]net.Interface, error) {
	var resultIfaces []net.Interface
	var errs error
	for _, interfaceIdentifier := range interfaces {
		if interfaceIdentifier == "" {
			continue
		}
		var iface *net.Interface
		var err error
		iface, err = net.InterfaceByName(interfaceIdentifier)
		if err != nil {
			ifaceIndex, err := strconv.Atoi(interfaceIdentifier)
			if err != nil {
				errs = err
				continue
			}
			iface, err = net.InterfaceByIndex(ifaceIndex)
			if err != nil {
				errs = err
				continue
			}
		}

		resultIfaces = append(resultIfaces, *iface)
	}

	if errs != nil {
		return resultIfaces, errs
	}

	return resultIfaces, nil
}
