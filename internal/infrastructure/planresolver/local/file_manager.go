package local

import (
	"context"
	"insync/internal/domain"
)

type IFileManager interface {
	GetSnapshot(ctx context.Context, rootName domain.RootName) (domain.SnapshotWithMetadata, error)
}
