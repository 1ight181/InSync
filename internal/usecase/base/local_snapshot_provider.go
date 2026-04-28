package base

import (
	"context"
	"insync/internal/domain"
)

type ILocalSnapshotProvider interface {
	GetLocalSnapshot(ctx context.Context, rootName domain.RootName) (snapshot domain.Snapshot, err error)
}
