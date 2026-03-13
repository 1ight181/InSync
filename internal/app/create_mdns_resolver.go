package app

import (
	"context"
	mdnsifaces "insync/internal/mdns/interfaces"
	"log/slog"

	mdns "insync/internal/mdns"
)

func createMDnsResolver(
	serviceType string,
	domain string,
	interfaces []string,
	logger *slog.Logger,
	ctx context.Context,
) (mdnsifaces.IResolver, error) {
	mDnsResolverOpts := mdns.MDnsResolverOptions{
		ServiceType: serviceType,
		Domain:      domain,
		Interfaces:  interfaces,
		Logger:      logger,
		Ctx:         ctx,
	}

	return mdns.NewMDnsResolver(mDnsResolverOpts), nil
}
