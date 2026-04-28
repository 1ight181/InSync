package local

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type LocalSnapshotProvider struct {
	fileManager          IFileManager
	baseSnapshotProvider IBaseSnapshotProvider
}

type LocalSnapshotProviderOptions struct {
	FileManager          IFileManager
	BaseSnapshotProvider IBaseSnapshotProvider
}

var (
	ErrInvalidOpts = errors.New("Все поля LocalSnapshotProviderOptions должны быть заполнены")
)

func NewLocalSnapshotProvider(opts LocalSnapshotProviderOptions) (*LocalSnapshotProvider, error) {
	if opts.FileManager == nil ||
		opts.BaseSnapshotProvider == nil {
		return nil, ErrInvalidOpts
	}
	return &LocalSnapshotProvider{
		fileManager:          opts.FileManager,
		baseSnapshotProvider: opts.BaseSnapshotProvider,
	}, nil
}

func (p *LocalSnapshotProvider) GetLocalSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error) {
	var baseSnapshot *domain.Snapshot
	rawBaseSnapshot, err := p.baseSnapshotProvider.GetBaseSnapshot(ctx, rootName)
	if err != nil && !errors.Is(err, domain.ErrBaseSnapshotNotFound) {
		return domain.Snapshot{}, err
	}

	if errors.Is(err, domain.ErrBaseSnapshotNotFound) {
		baseSnapshot = nil
	} else {
		baseSnapshot = &rawBaseSnapshot
	}

	snapshot, err := p.fileManager.GetSnapshot(ctx, rootName, baseSnapshot)
	if err != nil {
		return domain.Snapshot{}, err
	}

	return snapshot, nil
}
