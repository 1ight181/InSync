package syncer

import (
	"context"
	"insync/internal/domain"
)

type IPostSyncBaseSnapshotPersister interface {
	UpdateBaseSnapshot(ctx context.Context, rootName domain.RootName) error
}
