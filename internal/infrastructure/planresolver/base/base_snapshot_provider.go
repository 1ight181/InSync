package base

import (
	"context"
	"insync/internal/domain"
)

type BaseSnapshotProvider struct {
	baseSnapshotRepository IBaseSnapshotRepositoryReader
	deviceIdProvider       IDeviceIdProvider
}

type BaseSnapshotProviderOptions struct {
	BaseSnapshotRepository IBaseSnapshotRepositoryReader
}

func NewBaseSnapshotProvider(opts BaseSnapshotProviderOptions) *BaseSnapshotProvider {
	if opts.BaseSnapshotRepository == nil {
		panic("Все поля BaseSnapshotProviderOptions должны быть заполнены")
	}
	return &BaseSnapshotProvider{baseSnapshotRepository: opts.BaseSnapshotRepository}
}

func (b *BaseSnapshotProvider) GetSnapshot(ctx context.Context, rootName string) (domain.Snapshot, error) {
	remoteDeviceId, err := b.deviceIdProvider.GetCurrentRemoteDeviceId()
	if err != nil {
		return domain.Snapshot{}, err
	}

	localDeviceId := b.deviceIdProvider.GetCurrentLocalDeviceId()

	return b.baseSnapshotRepository.GetLastBaseSnapshotByDeviceIdAndRootName(
		ctx,
		localDeviceId, remoteDeviceId,
		rootName,
	)
}
