package app

import (
	"insync/internal/domain"
	"insync/internal/infrastructure/config/models"
	conn "insync/internal/infrastructure/connection"
	"insync/internal/infrastructure/deviceid"
	deviceidlocal "insync/internal/infrastructure/deviceid/local"
	deviceidcreator "insync/internal/infrastructure/deviceid/local/creator"
	deviceidremote "insync/internal/infrastructure/deviceid/remote"
	"insync/internal/infrastructure/filesys"
	"log/slog"
	"os"
)

func createDeviceIdProvider(
	deviceIdProviderLogger *slog.Logger,
	connectionManager *conn.ConnectionManager,
	deviceIdResolverConfig models.DeviceIdResolverConfig,
	fileSystem *filesys.FileSystem,
) (*deviceid.DeviceIdProvider, error) {
	remoteDeviceIdResolver, err := deviceidremote.NewRemoteDeviceIdResolver(connectionManager)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(deviceIdResolverConfig.DeviceIdDir, 0755); err != nil {
		return nil, err
	}

	deviceIdFilePath, err := domain.NewPath(deviceIdResolverConfig.GetDeviceIdFilePath())
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(deviceIdResolverConfig.DeviceIdDir, 0755); err != nil {
		return nil, err
	}

	deviceIdLocalCreatorOpts := deviceidcreator.LocalDeviceIdCreatorOptions{
		FileSys:          fileSystem,
		DeviceIdFilePath: deviceIdFilePath,
	}

	deviceidlocalCreator, err := deviceidcreator.NewLocalDeviceIdCreator(deviceIdLocalCreatorOpts)
	if err != nil {
		return nil, err
	}

	localIdDeviceResolver, err := deviceidlocal.NewLocalDeviceIdResolver(deviceidlocalCreator)
	if err != nil {
		return nil, err
	}

	deviceIdProviderOpts := deviceid.DeviceIdProviderOptions{
		LocalDeviceIdResolver:  localIdDeviceResolver,
		RemoteDeviceIdResolver: remoteDeviceIdResolver,
	}

	deviceIdProvider, err := deviceid.NewDeviceIdProvider(deviceIdProviderOpts)
	if err != nil {
		return nil, err
	}

	return deviceIdProvider, nil
}
