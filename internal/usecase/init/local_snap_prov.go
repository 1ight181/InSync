package init

import (
	"context"
	"insync/internal/domain"
)

type ILocalSnapshotProvider interface {
	GetLocalSnapshot(ctx context.Context, rootName domain.RootName, baseSnapshot *domain.BaseSnapshot) (domain.Snapshot, error)
}
