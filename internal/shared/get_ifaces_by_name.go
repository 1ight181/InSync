package resolver

import (
	"net"

	"go.uber.org/multierr"
)

func GetNetworkInterfacesByName(interfaceNames []string) ([]net.Interface, error) {
	var interfaces []net.Interface
	var errs error
	for _, interfaceName := range interfaceNames {
		if interfaceName == "" {
			continue
		}
		iface, err := net.InterfaceByName(interfaceName)
		if err != nil {
			errs = multierr.Append(
				errs,
				FailedToGetNetworkInterfaceByNameError{
					Err:           err,
					InterfaceName: interfaceName,
				},
			)
			continue
		}

		interfaces = append(interfaces, *iface)
	}

	if errs != nil {
		return interfaces, errs
	}

	return interfaces, nil
}
