package scan

import (
	"context"
	"insync/internal/domain"
)

type ScanUseCase struct {
	clientFabric   IClientFabric
	fileManager    IFileManager
	changesScanner IChangesScanner
}

type ScanUseCaseOptions struct {
	ClientFabric   IClientFabric
	ChangesScanner IChangesScanner
}

func NewScanUseCase(opts ScanUseCaseOptions) *ScanUseCase {
	if opts.ClientFabric == nil ||
		opts.ChangesScanner == nil {
		panic("Все поля ScanUseCaseOptions должны быть заполнены")
	}
	return &ScanUseCase{
		clientFabric: opts.ClientFabric,
	}
}

func (s *ScanUseCase) PlanSyncChanges(ctx context.Context, rootName domain.RootName) ([]domain.SyncChange, error) {
	localSnapshot, err := s.fileManager.GetSnapshot(ctx, rootName)
	if err != nil {
		return nil, err
	}

	remoteSnapshot, err := s.clientFabric.CurrentClient().GetSnapshot(ctx, rootName)
	if err != nil {
		return nil, err
	}

	return s.changesScanner.Scan(localSnapshot, remoteSnapshot)
}
