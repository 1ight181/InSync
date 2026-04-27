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

func NewLocalSnapshotProvider(opts LocalSnapshotProviderOptions) (*LocalSnapshotProvider, error) {
	if opts.FileManager == nil {
		return nil, ErrInvalidOpts
	}
	return &LocalSnapshotProvider{fileManager: opts.FileManager}, nil
}

func (p *LocalSnapshotProvider) GetLocalSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error) {
	snapshot, err := p.fileManager.GetSnapshot(ctx, rootName)
	if err != nil {
		return domain.Snapshot{}, err
	}

	return snapshot, nil
}
