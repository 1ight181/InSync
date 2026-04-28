package server

import (
	"context"
	"insync/internal/domain"
)

type IBaseSnapshotUseCase interface {
	UpdateBaseSnapshot(ctx context.Context, rootName domain.RootName) error
}
