package app

import (
	mdns "insync/internal/infrastructure/mdns"
	"log/slog"
)

func startMDnsServer(
	instanceName,
	serviceType,
	domain string,
	port int,
	interfaces []string,
	logger *slog.Logger,
) error {
	mDnsServerOpts := mdns.MDnsServerOptions{
		InstanceName: instanceName,
		ServiceType:  serviceType,
		Domain:       domain,
		Port:         port,
		Interfaces:   interfaces,
		Logger:       logger,
	}

	mDnsServer := mdns.NewMDnsServer(mDnsServerOpts)
	if err := mDnsServer.Start(); err != nil {
		return err
	}

	return nil
}
