package base

import (
	"context"
	"insync/internal/domain"
)

type IBaseSnapshotRepositoryReader interface {
	GetLastBaseSnapshotByDeviceIdAndRootName(ctx context.Context, localDeviceId, remoteDeviceId domain.DeviceId, rootName domain.RootName) (domain.BaseSnapshot, error)
	CreateBaseSnapshot(ctx context.Context, baseSnapshot domain.Snapshot, isInitial bool, localDeviceId, remoteDeviceId domain.DeviceId, rootName domain.RootName) error
}
