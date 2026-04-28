package server

import (
	"context"
	"insync/internal/domain"
)

type IBaseSnapshotUseCase interface {
	SetBaseSnapshot(ctx context.Context, snapshot domain.Snapshot, rootName domain.RootName) error
}
