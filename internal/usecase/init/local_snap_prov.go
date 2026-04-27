package init

import (
	"context"
	"insync/internal/domain"
)

type ILocalSnapshotProvider interface {
	GetLocalSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error)
}
