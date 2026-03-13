package mdns

import (
	mdnserr "insync/internal/mdns/errors"
	"net"

	"go.uber.org/multierr"
)

func getNetworkInterfacesByName(interfaceNames []string) ([]net.Interface, error) {
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
				mdnserr.FailedToGetNetworkInterfaceByNameError{
					Err:       err,
					Interface: interfaceName,
				},
			)
			continue
		}

		interfaces = append(interfaces, *iface)
		rawAddresses, err := iface.Addrs()
		if err != nil {
			errs = multierr.Append(
				errs,
				mdnserr.FailedToGetAddressesError{
					Err:       err,
					Interface: interfaceName,
				},
			)
			continue
		}

		var addresses []string
		for _, rawAddress := range rawAddresses {
			addresses = append(addresses, rawAddress.String())
		}
	}

	if errs != nil {
		return nil, errs
	}

	return interfaces, nil
}
