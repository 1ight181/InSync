package cli

import (
	"context"
	"insync/internal/domain"
)

type ISyncUseCase interface {
	ApplySyncChanges(ctx context.Context, shouldUseCache bool, rootName domain.RootName) (<-chan domain.ChangeEvent, error)
}
