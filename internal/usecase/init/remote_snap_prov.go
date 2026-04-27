package init

import (
	"context"
	"insync/internal/domain"
)

type IRemoteSnapshotProvider interface {
	GetRemoteSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error)
}
