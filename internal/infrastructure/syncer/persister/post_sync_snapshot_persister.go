package base

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type PostSyncBaseSnapshotPersister struct {
	baseSnapshotManager   IBaseSnapshotManager
	localSnapshotProvider ILocalSnapshotProvider
	clientFactory         IClientFactory
}

type PostSyncBaseSnapshotPersisterOptions struct {
	BaseSnapshotManager   IBaseSnapshotManager
	LocalSnapshotProvider ILocalSnapshotProvider
	ClientFactory         IClientFactory
}

var (
	ErrInvalidOpts = errors.New("Все поля PostSyncBaseSnapshotPersister не должны быть nil")
)

func NewPostSyncBaseSnapshotPersister(opts PostSyncBaseSnapshotPersisterOptions) (*PostSyncBaseSnapshotPersister, error) {
	if opts.BaseSnapshotManager == nil ||
		opts.LocalSnapshotProvider == nil ||
		opts.ClientFactory == nil {
		return nil, ErrInvalidOpts
	}
	return &PostSyncBaseSnapshotPersister{
		baseSnapshotManager:   opts.BaseSnapshotManager,
		localSnapshotProvider: opts.LocalSnapshotProvider,
		clientFactory:         opts.ClientFactory,
	}, nil
}

func (b *PostSyncBaseSnapshotPersister) UpdateBaseSnapshot(ctx context.Context, rootName domain.RootName) error {
	if err := b.updateBaseSnapshotLocaly(ctx, rootName); err != nil {
		return err
	}

	return b.updateBaseSnapshotRemotly(ctx, rootName)

}

func (b *PostSyncBaseSnapshotPersister) updateBaseSnapshotLocaly(ctx context.Context, rootName domain.RootName) error {
	currentBaseSnapshot, err := b.baseSnapshotManager.GetBaseSnapshot(ctx, rootName)
	if err != nil {
		return err
	}
	newBaseSnapshot, err := b.localSnapshotProvider.GetLocalSnapshot(ctx, rootName, &currentBaseSnapshot)
	if err != nil {
		return err
	}

	return b.baseSnapshotManager.CreateBaseSnapshot(ctx, newBaseSnapshot, rootName)
}

func (b *PostSyncBaseSnapshotPersister) updateBaseSnapshotRemotly(ctx context.Context, rootName domain.RootName) error {
	client, err := b.clientFactory.CurrentClient()
	if err != nil {
		return err
	}

	return client.UpdateBaseSnapshot(ctx, rootName)
}
