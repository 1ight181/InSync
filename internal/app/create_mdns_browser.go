package app

import (
	"context"
	mdns "insync/internal/infrastructure/mdns"
	"log/slog"
)

func createMDnsNodeNamesBrowser(
	serverServiceType string,
	serverDomain string,
	interfaces []string,
	logger *slog.Logger,
	ctx context.Context,
) (*mdns.MDnsNodeNamesBrowser, error) {
	mDnsResolverOpts := mdns.MDnsNodeNamesBrowserOptions{
		ServerServiceType: serverServiceType,
		ServerDomain:      serverDomain,
		Interfaces:        interfaces,
		Logger:            logger,
		Ctx:               ctx,
	}

	return mdns.NewMDnsNodeNamesBrowser(mDnsResolverOpts), nil
}
