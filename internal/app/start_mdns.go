package app

import (
	"context"
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
	ctx context.Context,
) error {
	mDnsServerOpts := mdns.MDnsServerOptions{
		InstanceName: instanceName,
		ServiceType:  serviceType,
		Domain:       domain,
		Port:         port,
		Interfaces:   interfaces,
		Logger:       logger,
		Ctx:          ctx,
	}

	mDnsServer := mdns.NewMDnsServer(mDnsServerOpts)
	if err := mDnsServer.Start(); err != nil {
		return err
	}

	return nil
}
