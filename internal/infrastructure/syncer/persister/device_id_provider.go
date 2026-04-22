package base

import "insync/internal/domain"

type IDeviceIdProvider interface {
	GetCurrentLocalDeviceId() (domain.DeviceId, error)
	GetCurrentRemoteDeviceId() (domain.DeviceId, error)
}
