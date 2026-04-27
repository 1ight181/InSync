package base

import (
	"context"
	"insync/internal/domain"
)

type IBaseSnapshotRepositoryReader interface {
	GetLastBaseSnapshotByDeviceIdAndRootName(ctx context.Context, localDeviceId, remoteDeviceId domain.DeviceId, rootName domain.RootName) (domain.Snapshot, error)
	CreateBaseSnapshot(ctx context.Context, baseSnapshot domain.Snapshot, localDeviceId, remoteDeviceId domain.DeviceId, rootName domain.RootName) error
}
