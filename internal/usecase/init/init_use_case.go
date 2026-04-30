package init

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type InitUseCase struct {
	baseSnapshotManager    IBaseSnapshotManager
	remoteSnapshotProvider IRemoteSnapshotProvider
	localSnapshotProvider  ILocalSnapshotProvider
}

type InitUseCaseOptions struct {
	BaseSnapshotManager    IBaseSnapshotManager
	RemoteSnapshotProvider IRemoteSnapshotProvider
	LocalSnapshotProvider  ILocalSnapshotProvider
}

var (
	ErrInvalidInitUseCaseOptions = errors.New("Все поля InitUseCaseOptions должны быть заполнены")
)

func NewInitUseCase(opts InitUseCaseOptions) (*InitUseCase, error) {
	if opts.BaseSnapshotManager == nil ||
		opts.RemoteSnapshotProvider == nil ||
		opts.LocalSnapshotProvider == nil {
		return nil, ErrInvalidInitUseCaseOptions
	}

	return &InitUseCase{
		baseSnapshotManager:    opts.BaseSnapshotManager,
		remoteSnapshotProvider: opts.RemoteSnapshotProvider,
		localSnapshotProvider:  opts.LocalSnapshotProvider,
	}, nil
}

func (i *InitUseCase) InitFromLocal(ctx context.Context, rootName domain.RootName) error {
	currentBaseSnapshot, err := i.baseSnapshotManager.GetBaseSnapshot(ctx, rootName)
	if err != nil {
		return err
	}

	localBase, err := i.localSnapshotProvider.GetLocalSnapshot(ctx, rootName, &currentBaseSnapshot)
	if err != nil {
		return err
	}

	return i.baseSnapshotManager.CreateBaseSnapshot(ctx, localBase, rootName)
}

func (i *InitUseCase) InitFromRemote(ctx context.Context, rootName domain.RootName) error {
	remoteBase, err := i.remoteSnapshotProvider.GetRemoteSnapshot(ctx, rootName)
	if err != nil {
		return err
	}

	return i.baseSnapshotManager.CreateBaseSnapshot(ctx, remoteBase, rootName)
}
