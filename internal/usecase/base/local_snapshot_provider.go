package base

import (
	"context"
	"insync/internal/domain"
)

type ILocalSnapshotProvider interface {
	GetLocalSnapshot(ctx context.Context, rootName domain.RootName, baseSnapshot *domain.Snapshot) (snapshot domain.Snapshot, err error)
}
