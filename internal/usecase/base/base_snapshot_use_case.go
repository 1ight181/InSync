package base

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type BaseSnapshotUseCase struct {
	localSnapshotProvider ILocalSnapshotProvider
	baseSnapshotManager   IBaseSnapshotManager
}

type BaseSnapshotUseCaseOpts struct {
	LocalSnapshotProvider ILocalSnapshotProvider
	BaseSnapshotManager   IBaseSnapshotManager
}

var (
	ErrInvalidOpts = errors.New("Все поля BaseSnapshotUseCaseOptions должны быть заполнены")
)

func NewBaseSnapshotUseCase(opts BaseSnapshotUseCaseOpts) (*BaseSnapshotUseCase, error) {
	if opts.LocalSnapshotProvider == nil ||
		opts.BaseSnapshotManager == nil {
		return nil, ErrInvalidOpts
	}
	return &BaseSnapshotUseCase{
		localSnapshotProvider: opts.LocalSnapshotProvider,
		baseSnapshotManager:   opts.BaseSnapshotManager,
	}, nil
}

func (b *BaseSnapshotUseCase) UpdateBaseSnapshot(ctx context.Context, rootName domain.RootName) error {
	currentBaseSnapshot, err := b.baseSnapshotManager.GetBaseSnapshot(ctx, rootName)
	if err != nil {
		return err
	}

	localSnapshot, err := b.localSnapshotProvider.GetLocalSnapshot(ctx, rootName, &currentBaseSnapshot)
	if err != nil {
		return err
	}

	return b.baseSnapshotManager.CreateBaseSnapshot(ctx, localSnapshot, rootName)
}
