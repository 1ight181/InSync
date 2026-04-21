package deviceid

import (
	"insync/internal/domain"
)

type IRemoteDeviceIdResolver interface {
	Resolve() (domain.DeviceId, error)
}
