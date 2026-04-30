package init

import (
	"context"
	"insync/internal/domain"
)

type IBaseSnapshotManager interface {
	CreateBaseSnapshot(ctx context.Context, baseSnapshot domain.Snapshot, rootName domain.RootName) error
}
