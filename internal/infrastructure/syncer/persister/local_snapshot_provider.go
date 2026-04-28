package base

import (
	"context"
	"insync/internal/domain"
)

type ILocalSnapshotProvider interface {
	GetSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error)
}
