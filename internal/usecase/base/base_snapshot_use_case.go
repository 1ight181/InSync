package base

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type BaseSnapshotUseCase struct {
	localSnapshotProvider ILocalSnapshotProvider
	baseSnapshotCreator   IBaseSnapshotCreator
}

type BaseSnapshotUseCaseOpts struct {
	LocalSnapshotProvider ILocalSnapshotProvider
	BaseSnapshotCreator   IBaseSnapshotCreator
}

var (
	ErrInvalidOpts = errors.New("Все поля BaseSnapshotUseCaseOptions должны быть заполнены")
)

func NewBaseSnapshotUseCase(opts BaseSnapshotUseCaseOpts) (*BaseSnapshotUseCase, error) {
	if opts.LocalSnapshotProvider == nil ||
		opts.BaseSnapshotCreator == nil {
		return nil, ErrInvalidOpts
	}
	return &BaseSnapshotUseCase{
		localSnapshotProvider: opts.LocalSnapshotProvider,
		baseSnapshotCreator:   opts.BaseSnapshotCreator,
	}, nil
}

func (b *BaseSnapshotUseCase) UpdateBaseSnapshot(ctx context.Context, baseSnapshot domain.Snapshot, rootName domain.RootName) error {
	localSnapshot, err := b.localSnapshotProvider.GetLocalSnapshot(ctx, rootName)
	if err != nil {
		return err
	}

	return b.baseSnapshotCreator.CreateBaseSnapshot(ctx, localSnapshot, rootName)
}
