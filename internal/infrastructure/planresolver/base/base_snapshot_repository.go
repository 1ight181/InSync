package base

import (
	"context"
	"insync/internal/domain"
)

type IBaseSnapshotRepositoryReader interface {
	GetLastBaseSnapshotByDeviceIdAndRootName(ctx context.Context, localDeviceId, remoteDeviceId string, rootName domain.RootName) (domain.Snapshot, error)
}

type IBaseSnapshotRepositoryWriter interface {
	CreateBaseSnapshot(ctx context.Context, baseSnapshot domain.Snapshot) error
}
