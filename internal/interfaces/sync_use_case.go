package interfaces

import "insync/internal/domain"

type ISyncUseCase interface {
	GetSyncChanges() []domain.SyncChange
	ApplySyncChanges(changes []domain.SyncChange) error
}
