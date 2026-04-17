package deviceid

import "insync/internal/domain"

type IRemoteDeviceIdResolver interface {
	Resolve(nodeName domain.NodeName) (domain.DeviceId, error)
}
