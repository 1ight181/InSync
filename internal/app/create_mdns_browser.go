package app

import (
	"insync/internal/infrastructure/config/models"
	mdns "insync/internal/infrastructure/mdns"
	"log/slog"
)

func createMDnsNodeNamesBrowser(
	mDnsBrowserLogger *slog.Logger,
	mDnsBrowserConfig models.MDnsBrowserConfig,
) (*mdns.MDnsNodeNamesBrowser, error) {
	mDnsResolverOpts := mdns.MDnsNodeNamesBrowserOptions{
		ServerServiceType: mDnsBrowserConfig.ServerServiceType,
		ServerDomain:      mDnsBrowserConfig.ServerDomain,
		Interfaces:        mDnsBrowserConfig.GetInterfaces(),
		Logger:            mDnsBrowserLogger,
	}

	return mdns.NewMDnsNodeNamesBrowser(mDnsResolverOpts), nil
}
