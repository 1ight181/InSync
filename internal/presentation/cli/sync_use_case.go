package cli

import "insync/internal/domain"

type ISyncUseCase interface {
	GetSyncChanges(rootName domain.RootName) ([]domain.SyncChange, error)
	ApplySyncChanges(changes []domain.SyncChange) (<-chan domain.ChangeEvent, error)
}
