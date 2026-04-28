package base

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type BaseSnapshotManager struct {
	baseSnapshotRepository IBaseSnapshotRepositoryReader
	deviceIdProvider       IDeviceIdProvider
}

type BaseSnapshotManagerOptions struct {
	BaseSnapshotRepository IBaseSnapshotRepositoryReader
	DeviceIdProvider       IDeviceIdProvider
}

var (
	ErrInvalidOpts = errors.New("Все поля BaseSnapshotManagerOptions должны быть заполнены")
)

func NewBaseSnapshotManager(opts BaseSnapshotManagerOptions) (*BaseSnapshotManager, error) {
	if opts.BaseSnapshotRepository == nil ||
		opts.DeviceIdProvider == nil {
		return nil, ErrInvalidOpts
	}
	return &BaseSnapshotManager{
		baseSnapshotRepository: opts.BaseSnapshotRepository,
		deviceIdProvider:       opts.DeviceIdProvider,
	}, nil
}

func (b *BaseSnapshotManager) GetBaseSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error) {

	localDeviceId, err := b.deviceIdProvider.GetCurrentLocalDeviceId()
	if err != nil {
		return domain.Snapshot{}, err
	}

	remoteDeviceId, err := b.deviceIdProvider.GetCurrentRemoteDeviceId()
	if err != nil {
		return domain.Snapshot{}, err
	}

	return b.baseSnapshotRepository.GetLastBaseSnapshotByDeviceIdAndRootName(
		ctx,
		localDeviceId, remoteDeviceId,
		rootName,
	)
}

func (b *BaseSnapshotManager) CreateBaseSnapshot(ctx context.Context, baseSnapshot domain.Snapshot, rootName domain.RootName) error {
	localDeviceId, err := b.deviceIdProvider.GetCurrentLocalDeviceId()
	if err != nil {
		return err
	}

	remoteDeviceId, err := b.deviceIdProvider.GetCurrentRemoteDeviceId()
	if err != nil {
		return err
	}

	return b.baseSnapshotRepository.CreateBaseSnapshot(ctx, baseSnapshot, localDeviceId, remoteDeviceId, rootName)
}
