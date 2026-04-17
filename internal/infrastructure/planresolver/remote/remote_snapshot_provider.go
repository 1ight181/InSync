package remote

import (
	"context"
	"insync/internal/domain"
)

type RemoteSnapshotProvider struct {
	clientFabric IClientFabric
}

type RemoteSnapshotProviderOptions struct {
	ClientFabric IClientFabric
}

func NewRemoteSnapshotProvider(options RemoteSnapshotProviderOptions) *RemoteSnapshotProvider {
	if options.ClientFabric == nil {
		panic("Все поля RemoteSnapshotProviderOptions должны быть заполнены")
	}
	return &RemoteSnapshotProvider{clientFabric: options.ClientFabric}
}

func (p *RemoteSnapshotProvider) GetRemoteSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error) {
	client := p.clientFabric.CurrentClient()
	snapshotWithMetadata, err := client.GetSnapshot(ctx, rootName)
	if err != nil {
		return domain.Snapshot{}, err
	}

	return snapshotWithMetadata.Snapshot, nil
}
