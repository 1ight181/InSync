package base

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type BaseSnapshotProvider struct {
	baseSnapshotRepository IBaseSnapshotRepositoryReader
	deviceIdProvider       IDeviceIdProvider
}

type BaseSnapshotProviderOptions struct {
	BaseSnapshotRepository IBaseSnapshotRepositoryReader
	DeviceIdProvider       IDeviceIdProvider
}

var (
	ErrInvalidOpts = errors.New("Все поля BaseSnapshotProviderOptions должны быть заполнены")
)

func NewBaseSnapshotProvider(opts BaseSnapshotProviderOptions) (*BaseSnapshotProvider, error) {
	if opts.BaseSnapshotRepository == nil ||
		opts.DeviceIdProvider == nil {
		return nil, ErrInvalidOpts
	}
	return &BaseSnapshotProvider{
		baseSnapshotRepository: opts.BaseSnapshotRepository,
		deviceIdProvider:       opts.DeviceIdProvider,
	}, nil
}

func (b *BaseSnapshotProvider) GetBaseSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error) {
	remoteDeviceId, err := b.deviceIdProvider.GetCurrentRemoteDeviceId()
	if err != nil {
		return domain.Snapshot{}, err
	}

	localDeviceId, err := b.deviceIdProvider.GetCurrentLocalDeviceId()
	if err != nil {
		return domain.Snapshot{}, err
	}

	return b.baseSnapshotRepository.GetLastBaseSnapshotByDeviceIdAndRootName(
		ctx,
		localDeviceId, remoteDeviceId,
		rootName,
	)
}
