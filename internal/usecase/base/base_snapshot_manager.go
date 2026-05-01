package base

import (
	"context"
	"insync/internal/domain"
)

type IBaseSnapshotManager interface {
	CreateBaseSnapshot(ctx context.Context, baseSnapshot domain.Snapshot, rootName domain.RootName, isInitial bool) error
	GetBaseSnapshot(ctx context.Context, rootName domain.RootName) (domain.BaseSnapshot, error)
	DeleteBaseSnapshots(ctx context.Context, rootName domain.RootName) error
}
