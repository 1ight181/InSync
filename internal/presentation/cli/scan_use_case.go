package cli

import "insync/internal/domain"

type IScanUseCase interface {
	PlanSyncChanges(rootName domain.RootName) ([]domain.SyncChange, error)
}
