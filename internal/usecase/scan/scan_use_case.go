package sync

import "insync/internal/domain"

type SyncUseCase struct {
	clientFabric   IClientFabric
	fileManager    IFileManager
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

func (s *SyncUseCase) PlanSyncChanges(rootName domain.RootName) ([]domain.SyncChange, error) {
	localEntries, err := s.fileManager.GetFileList(rootName)
	if err != nil {
		return nil, err
	}

	remoteEntries, err := s.clientFabric.CurrentClient().GetFileList(rootName.String())
	if err != nil {
		return nil, err
	}

	return s.changesScanner.Scan(localEntries, remoteEntries)
}
