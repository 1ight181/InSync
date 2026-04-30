package local

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type LocalSnapshotProvider struct {
	fileManager IFileManager
}

type LocalSnapshotProviderOptions struct {
	FileManager IFileManager
}

var (
	ErrInvalidOpts = errors.New("Все поля LocalSnapshotProviderOptions должны быть заполнены")
)

func NewLocalSnapshotProvider(fileManager IFileManager) (*LocalSnapshotProvider, error) {
	if fileManager == nil {
		return nil, ErrInvalidOpts
	}
	return &LocalSnapshotProvider{
		fileManager: fileManager,
	}, nil
}

func (p *LocalSnapshotProvider) GetLocalSnapshot(ctx context.Context, rootName domain.RootName, baseSnapshot *domain.BaseSnapshot) (domain.Snapshot, error) {
	snapshot, err := p.fileManager.GetSnapshot(ctx, rootName, baseSnapshot)
	if err != nil {
		return domain.Snapshot{}, err
	}

	return snapshot, nil
}
