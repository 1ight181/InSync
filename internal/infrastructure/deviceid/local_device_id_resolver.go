package deviceid

import "insync/internal/domain"

type ILocalDeviceIdResolver interface {
	Resolve() domain.DeviceId
}
