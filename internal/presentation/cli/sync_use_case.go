package cli

import "insync/internal/domain"

type ISyncUseCase interface {
	GetSyncChanges(rootName domain.RootName) []domain.SyncChange
	ApplySyncChanges(changes []domain.SyncChange) (<-chan domain.ChangeEvent, error)
}
