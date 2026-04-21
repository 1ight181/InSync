package base

import "insync/internal/domain"

type IDeviceIdProvider interface {
	GetCurrentLocalDeviceId() domain.DeviceId
	GetCurrentRemoteDeviceId() (domain.DeviceId, error)
}
