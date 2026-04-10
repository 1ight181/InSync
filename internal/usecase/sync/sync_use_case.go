package sync

import "insync/internal/domain"

type SyncUseCase struct {
	clientFabric IClientFabric
}

type SyncUseCaseOptions struct {
	ClientFabric IClientFabric
}

func NewSyncUseCase(opts SyncUseCaseOptions) *SyncUseCase {
	if opts.ClientFabric == nil {
		panic("Все поля SyncUseCaseOptions должны быть заполнены")
	}
	return &SyncUseCase{
		clientFabric: opts.ClientFabric,
	}
}

func (s *SyncUseCase) ApplySyncChanges(changes []domain.SyncChange) (<-chan domain.ChangeEvent, error) {
	return make(<-chan domain.ChangeEvent), nil
}

func (s *SyncUseCase) GetSyncChanges(rootName domain.RootName) []domain.SyncChange {
	return []domain.SyncChange{}
}
