package sync

import (
	"context"
	"insync/internal/domain"
)

type ISyncer interface {
	Sync(ctx context.Context, plan domain.SyncPlan, rootName domain.RootName) (
		appliedChanges <-chan domain.ChangeEvent,
		conflicts <-chan domain.Conflict,
		userDecision chan<- domain.Decision,
		err error,
	)
}
