package base

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type PostSyncBaseSnapshotPersister struct {
	baseSnapshotRepository IBaseSnapshotRepositoryWriter
	deviceIdProvider       IDeviceIdProvider
	snapshotProvider       IBaseSnapshotProvider
}

type PostSyncBaseSnapshotPersisterOptions struct {
	BaseSnapshotRepository IBaseSnapshotRepositoryWriter
	DeviceIdProvider       IDeviceIdProvider
	SnapshotProvider       IBaseSnapshotProvider
}

var (
	ErrInvalidOpts = errors.New("Все поля PostSyncBaseSnapshotPersister должны быть заполнены")
)

func NewPostSyncBaseSnapshotPersister(opts PostSyncBaseSnapshotPersisterOptions) (*PostSyncBaseSnapshotPersister, error) {
	if opts.BaseSnapshotRepository == nil ||
		opts.DeviceIdProvider == nil ||
		opts.SnapshotProvider == nil {
		return nil, ErrInvalidOpts
	}
	return &PostSyncBaseSnapshotPersister{
		baseSnapshotRepository: opts.BaseSnapshotRepository,
		deviceIdProvider:       opts.DeviceIdProvider,
		snapshotProvider:       opts.SnapshotProvider,
	}, nil
}

func (b *PostSyncBaseSnapshotPersister) UpdateBaseSnapshot(ctx context.Context, rootName domain.RootName) error {
	newBaseSnapshot, err := b.snapshotProvider.GetBaseSnapshot(ctx, rootName)
	if err != nil {
		return err
	}

	remoteDeviceId, err := b.deviceIdProvider.GetCurrentRemoteDeviceId()
	if err != nil {
		return err
	}

	localDeviceId, err := b.deviceIdProvider.GetCurrentLocalDeviceId()
	if err != nil {
		return err
	}

	return b.baseSnapshotRepository.CreateBaseSnapshot(ctx, newBaseSnapshot, localDeviceId, remoteDeviceId, rootName)
}
