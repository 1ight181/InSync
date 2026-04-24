package local

import "insync/internal/domain"

type ILocalDeviceIdCreator interface {
	ReadOrCreateLocalDeviceId() (domain.DeviceId, error)
}
