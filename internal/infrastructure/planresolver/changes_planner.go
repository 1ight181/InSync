package sync

import (
	"context"
	"insync/internal/domain"
)

type IChangesPlanner interface {
	Plan(ctx context.Context, baseSnapshot, localSnapshot, remoteSnapshot domain.Snapshot) (domain.SyncPlan, error)
}
