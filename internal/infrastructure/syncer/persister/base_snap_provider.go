package base

import (
	"context"
	"insync/internal/domain"
)

type IBaseSnapshotProvider interface {
	GetBaseSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error)
}
