package local

import "insync/internal/domain"

type ILocalDeviceIdCreator interface {
	CreateLocalDeviceId() (domain.DeviceId, error)
}
