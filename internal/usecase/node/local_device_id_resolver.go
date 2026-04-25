package connect

import "insync/internal/domain"

type ILocalDeviceIdResolver interface {
	Resolve() (deviceId domain.DeviceId, err error)
}
