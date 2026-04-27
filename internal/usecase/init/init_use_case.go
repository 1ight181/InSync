package init

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type InitUseCase struct {
	baseSnapshotCreator    IBaseSnapshotCreator
	remoteSnapshotProvider IRemoteSnapshotProvider
	localSnapshotProvider  ILocalSnapshotProvider
}

type InitUseCaseOptions struct {
	BaseSnapshotCreator    IBaseSnapshotCreator
	RemoteSnapshotProvider IRemoteSnapshotProvider
	LocalSnapshotProvider  ILocalSnapshotProvider
}

var (
	ErrInvalidInitUseCaseOptions = errors.New("Все поля InitUseCaseOptions должны быть заполнены")
)

func NewInitUseCase(opts InitUseCaseOptions) (*InitUseCase, error) {
	if opts.BaseSnapshotCreator == nil ||
		opts.RemoteSnapshotProvider == nil ||
		opts.LocalSnapshotProvider == nil {
		return nil, ErrInvalidInitUseCaseOptions
	}

	return &InitUseCase{
		baseSnapshotCreator:    opts.BaseSnapshotCreator,
		remoteSnapshotProvider: opts.RemoteSnapshotProvider,
		localSnapshotProvider:  opts.LocalSnapshotProvider,
	}, nil
}

func (i *InitUseCase) InitFromLocal(ctx context.Context, rootName domain.RootName) error {
	localBase, err := i.localSnapshotProvider.GetLocalSnapshot(ctx, rootName)
	if err != nil {
		return err
	}

	return i.baseSnapshotCreator.CreateBaseSnapshot(ctx, localBase, rootName)
}

func (i *InitUseCase) InitFromRemote(ctx context.Context, rootName domain.RootName) error {
	remoteBase, err := i.remoteSnapshotProvider.GetRemoteSnapshot(ctx, rootName)
	if err != nil {
		return err
	}

	return i.baseSnapshotCreator.CreateBaseSnapshot(ctx, remoteBase, rootName)
}
