package base

import (
	"context"
	"insync/internal/domain"
)

type IBaseSnapshotRepositoryWriter interface {
	CreateBaseSnapshot(ctx context.Context, baseSnapshot domain.Snapshot, localDeviceId, remoteDeviceId domain.DeviceId) error
}
