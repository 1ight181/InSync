package app

import (
	mdns "insync/internal/infrastructure/mdns"
	"log/slog"
)

func createMDnsNodeNamesBrowser(
	serverServiceType string,
	serverDomain string,
	interfaces []string,
	logger *slog.Logger,
) (*mdns.MDnsNodeNamesBrowser, error) {
	mDnsResolverOpts := mdns.MDnsNodeNamesBrowserOptions{
		ServerServiceType: serverServiceType,
		ServerDomain:      serverDomain,
		Interfaces:        interfaces,
		Logger:            logger,
	}

	return mdns.NewMDnsNodeNamesBrowser(mDnsResolverOpts), nil
}
