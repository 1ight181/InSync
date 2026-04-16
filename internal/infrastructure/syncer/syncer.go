package syncer

import "insync/internal/domain"

type Syncer struct{}

func NewSyncer() *Syncer {
	return &Syncer{}
}

func (s *Syncer) Sync(plan domain.SyncPlan) (<-chan domain.ChangeEvent, error) {

	return make(chan domain.ChangeEvent), nil
}
