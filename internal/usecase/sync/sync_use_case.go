package sync

import "insync/internal/domain"

type SyncUseCase struct {
	clientFabric   IClientFabric
	changesScanner IChangesScanner
}

type SyncUseCaseOptions struct {
	ClientFabric   IClientFabric
	ChangesScanner IChangesScanner
}

func NewSyncUseCase(opts SyncUseCaseOptions) *SyncUseCase {
	if opts.ClientFabric == nil ||
		opts.ChangesScanner == nil {
		panic("Все поля SyncUseCaseOptions должны быть заполнены")
	}
	return &SyncUseCase{
		clientFabric: opts.ClientFabric,
	}
}

func (s *SyncUseCase) ApplySyncChanges(changes []domain.SyncChange) (<-chan domain.ChangeEvent, error) {
	return make(<-chan domain.ChangeEvent), nil
}

func (s *SyncUseCase) GetSyncChanges(rootName domain.RootName) ([]domain.SyncChange, error) {
	return s.changesScanner.Scan(rootName)
}
