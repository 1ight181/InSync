package cli

import (
	"context"
	"insync/internal/domain"
)

type IScanUseCase interface {
	PlanSyncChanges(ctx context.Context, rootName domain.RootName) ([]domain.SyncChange, error)
}
