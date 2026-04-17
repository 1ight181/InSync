package base

import (
	"context"
	"insync/internal/domain"
)

type IBaseSnapshotRepository interface {
	GetLastBaseSnapshotByDeviceIdAndRootName(ctx context.Context, localDeviceId, remoteDeviceId, rootName string) (domain.Snapshot, error)
	CreateBaseSnapshot(ctx context.Context, baseSnapshot domain.Snapshot) error
}
