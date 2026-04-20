package cli

import (
	"context"
	"insync/internal/domain"
)

type ISyncUseCase interface {
	ApplySyncChanges(ctx context.Context, shouldUseCache bool, rootName domain.RootName) (
		appliedChanges <-chan domain.ChangeEvent,
		conflicts <-chan domain.Conflict,
		userDecision chan<- domain.Decision,
		err error,
	)
}
