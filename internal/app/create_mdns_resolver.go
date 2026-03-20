package app

import (
	"context"
	mdnsifaces "insync/internal/interfaces"
	"log/slog"

	mdns "insync/internal/infrastructure/mdns"
)

func createMDnsBrowser(
	serviceType string,
	domain string,
	interfaces []string,
	logger *slog.Logger,
	ctx context.Context,
) (mdnsifaces.IMdnsBrowser, error) {
	mDnsResolverOpts := mdns.MDnsBrowserOptions{
		ServiceType: serviceType,
		Domain:      domain,
		Interfaces:  interfaces,
		Logger:      logger,
		Ctx:         ctx,
	}

	return mdns.NewMDnsBrowser(mDnsResolverOpts), nil
}
