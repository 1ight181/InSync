package sync

import "insync/internal/domain"

type ISyncer interface {
	Sync(plan domain.SyncPlan) (<-chan domain.ChangeEvent, error)
}
