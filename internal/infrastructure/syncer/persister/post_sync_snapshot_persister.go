package base

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type PostSyncBaseSnapshotPersister struct {
	baseSnapshotRepository IBaseSnapshotRepositoryWriter
	deviceIdProvider       IDeviceIdProvider
	localSnapshotProvider  ILocalSnapshotProvider
	clientFactory          IClientFactory
}

type PostSyncBaseSnapshotPersisterOptions struct {
	BaseSnapshotRepository IBaseSnapshotRepositoryWriter
	DeviceIdProvider       IDeviceIdProvider
	LocalSnapshotProvider  ILocalSnapshotProvider
	ClientFactory          IClientFactory
}

var (
	ErrInvalidOpts = errors.New("Все поля PostSyncBaseSnapshotPersister должны быть заполнены")
)

func NewPostSyncBaseSnapshotPersister(opts PostSyncBaseSnapshotPersisterOptions) (*PostSyncBaseSnapshotPersister, error) {
	if opts.BaseSnapshotRepository == nil ||
		opts.DeviceIdProvider == nil ||
		opts.LocalSnapshotProvider == nil ||
		opts.ClientFactory == nil {
		return nil, ErrInvalidOpts
	}
	return &PostSyncBaseSnapshotPersister{
		baseSnapshotRepository: opts.BaseSnapshotRepository,
		deviceIdProvider:       opts.DeviceIdProvider,
		localSnapshotProvider:  opts.LocalSnapshotProvider,
		clientFactory:          opts.ClientFactory,
	}, nil
}

func (b *PostSyncBaseSnapshotPersister) UpdateBaseSnapshot(ctx context.Context, rootName domain.RootName) error {
	if err := b.updateBaseSnapshotLocaly(ctx, rootName); err != nil {
		return err
	}

	return b.updateBaseSnapshotRemotly(ctx, rootName)

}

func (b *PostSyncBaseSnapshotPersister) updateBaseSnapshotLocaly(ctx context.Context, rootName domain.RootName) error {
	newBaseSnapshot, err := b.localSnapshotProvider.GetLocalSnapshot(ctx, rootName)
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

func (b *PostSyncBaseSnapshotPersister) updateBaseSnapshotRemotly(ctx context.Context, rootName domain.RootName) error {
	client, err := b.clientFactory.CurrentClient()
	if err != nil {
		return err
	}

	return client.UpdateBaseSnapshot(ctx, rootName)
}
