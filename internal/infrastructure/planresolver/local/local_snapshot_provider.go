package local

import (
	"context"
	"insync/internal/domain"
)

type LocalSnapshotProvider struct {
	fileManager IFileManager
}

type LocalSnapshotProviderOptions struct {
	FileManager IFileManager
}

func NewLocalSnapshotProvider(opts LocalSnapshotProviderOptions) *LocalSnapshotProvider {
	if opts.FileManager == nil {
		panic("Все поля LocalSnapshotProviderOptions должны быть заполнены")
	}
	return &LocalSnapshotProvider{fileManager: opts.FileManager}
}

func (p *LocalSnapshotProvider) GetLocalSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error) {
	snapshot, err := p.fileManager.GetSnapshot(ctx, rootName)
	if err != nil {
		return domain.Snapshot{}, err
	}

	return snapshot.Snapshot, nil
}
