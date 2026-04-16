package sync

import (
	"context"
	"insync/internal/domain"
)

type IPreviousSnapshotProvider interface {
	GetPreviousSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error)
}
