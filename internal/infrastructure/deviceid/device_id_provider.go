package deviceid

import (
	"errors"
	"insync/internal/domain"
)

type DeviceIdProvider struct {
	localDeviceIdResolver  ILocalDeviceIdResolver
	remoteDeviceIdResolver IRemoteDeviceIdResolver
}

type DeviceIdProviderOptions struct {
	LocalDeviceIdResolver  ILocalDeviceIdResolver
	RemoteDeviceIdResolver IRemoteDeviceIdResolver
}

var (
	ErrDeviceIdProviderInvalidOpts = errors.New("Все поля DeviceIdProviderOptions должны быть заполнены")
)

func NewDeviceIdProvider(opts DeviceIdProviderOptions) (*DeviceIdProvider, error) {
	if opts.LocalDeviceIdResolver == nil ||
		opts.RemoteDeviceIdResolver == nil {
		return nil, ErrDeviceIdProviderInvalidOpts
	}
	return &DeviceIdProvider{
		localDeviceIdResolver:  opts.LocalDeviceIdResolver,
		remoteDeviceIdResolver: opts.RemoteDeviceIdResolver,
	}, nil
}

func (d *DeviceIdProvider) GetCurrentLocalDeviceId() (domain.DeviceId, error) {
	return d.localDeviceIdResolver.Resolve()
}

func (d *DeviceIdProvider) GetCurrentRemoteDeviceId() (domain.DeviceId, error) {
	return d.remoteDeviceIdResolver.Resolve()
}
