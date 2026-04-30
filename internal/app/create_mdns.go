package app

import (
	"insync/internal/infrastructure/config/models"
	"insync/internal/infrastructure/deviceid"
	mdns "insync/internal/infrastructure/mdns"
	"log/slog"
)

func createMDnsServer(
	mDnsServerlogger *slog.Logger,
	mDnsServerConfig models.MDnsServerConfig,
	deviceIdProvider *deviceid.DeviceIdProvider,
) (*mdns.MDnsServer, error) {

	localDeviceId, err := deviceIdProvider.GetCurrentLocalDeviceId()
	if err != nil {
		return nil, err
	}

	mDnsServerInstanceName := localDeviceId.String()

	mDnsServerOpts := mdns.MDnsServerOptions{
		InstanceName: mDnsServerInstanceName,
		ServiceType:  mDnsServerConfig.ServiceType,
		Domain:       mDnsServerConfig.Domain,
		Port:         mDnsServerConfig.Port,
		Interfaces:   mDnsServerConfig.GetInterfaces(),
		Logger:       mDnsServerlogger,
	}

	mDnsServer, err := mdns.NewMDnsServer(mDnsServerOpts)
	if err != nil {
		return nil, err
	}

	return mDnsServer, nil
}
