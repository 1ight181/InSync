package local

import (
	"context"
	"insync/internal/domain"
)

type IFileManager interface {
	GetSnapshot(ctx context.Context, rootName domain.RootName, baseSnapshot *domain.BaseSnapshot) (domain.Snapshot, error)
}
