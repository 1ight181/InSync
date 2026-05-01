package init

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type InitUseCase struct {
	baseSnapshotManager   IBaseSnapshotManager
	localSnapshotProvider ILocalSnapshotProvider
}

type InitUseCaseOptions struct {
	BaseSnapshotManager   IBaseSnapshotManager
	LocalSnapshotProvider ILocalSnapshotProvider
}

var (
	ErrInvalidInitUseCaseOptions = errors.New("Все поля InitUseCaseOptions должны быть заполнены")
)

func NewInitUseCase(opts InitUseCaseOptions) (*InitUseCase, error) {
	if opts.BaseSnapshotManager == nil ||
		opts.LocalSnapshotProvider == nil {
		return nil, ErrInvalidInitUseCaseOptions
	}

	return &InitUseCase{
		baseSnapshotManager:   opts.BaseSnapshotManager,
		localSnapshotProvider: opts.LocalSnapshotProvider,
	}, nil
}

func (i *InitUseCase) Init(ctx context.Context, rootName domain.RootName) error {
	localBase, err := i.localSnapshotProvider.GetLocalSnapshot(ctx, rootName, nil)
	if err != nil {
		return err
	}

	return i.baseSnapshotManager.CreateBaseSnapshot(ctx, localBase, rootName, true)
}
