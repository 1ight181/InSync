package connect

import "insync/internal/domain"

type ILocalDeviceIdProvider interface {
	GetCurrentLocalDeviceId() (deviceId domain.DeviceId, err error)
}
