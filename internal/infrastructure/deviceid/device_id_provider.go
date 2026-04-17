package deviceid

import "insync/internal/domain"

type DeviceIdProvider struct {
	localDeviceIdResolver  ILocalDeviceIdResolver
	remoteDeviceIdResolver IRemoteDeviceIdResolver
}

type DeviceIdProviderOptions struct {
	LocalDeviceIdResolver  ILocalDeviceIdResolver
	RemoteDeviceIdResolver IRemoteDeviceIdResolver
}

func NewDeviceIdProvider(opts DeviceIdProviderOptions) *DeviceIdProvider {
	if opts.LocalDeviceIdResolver == nil ||
		opts.RemoteDeviceIdResolver == nil {
		panic("Все поля DeviceIdProviderOptions должны быть заполнены")
	}
	return &DeviceIdProvider{
		localDeviceIdResolver:  opts.LocalDeviceIdResolver,
		remoteDeviceIdResolver: opts.RemoteDeviceIdResolver,
	}
}

func (d *DeviceIdProvider) GetCurrentLocalDeviceId() domain.DeviceId {
	return d.localDeviceIdResolver.Resolve()
}

func (d *DeviceIdProvider) GetCurrentRemoteDeviceId(nodeName domain.NodeName) (domain.DeviceId, error) {
	remoteDeviceId, err := d.remoteDeviceIdResolver.Resolve(nodeName)
	if err != nil {
		return "", err
	}

	return remoteDeviceId, nil
}
