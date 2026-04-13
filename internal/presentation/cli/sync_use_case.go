package cli

import "insync/internal/domain"

type ISyncUseCase interface {
	ApplySyncChanges(changes []domain.SyncChange) (<-chan domain.ChangeEvent, error)
}
