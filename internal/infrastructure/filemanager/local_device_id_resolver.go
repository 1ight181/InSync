package filemanager

import (
	"insync/internal/domain"
)

type ILocalDeviceIdProvider interface {
	GetLocalDeviceId() domain.DeviceId
}
