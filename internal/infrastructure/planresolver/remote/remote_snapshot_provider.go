package remote

import (
	"context"
	"insync/internal/domain"
)

type RemoteSnapshotProvider struct {
	clientFactory IClientFactory
}

type RemoteSnapshotProviderOptions struct {
	ClientFactory IClientFactory
}

func NewRemoteSnapshotProvider(options RemoteSnapshotProviderOptions) *RemoteSnapshotProvider {
	if options.ClientFactory == nil {
		panic("Все поля RemoteSnapshotProviderOptions должны быть заполнены")
	}
	return &RemoteSnapshotProvider{clientFactory: options.ClientFactory}
}

func (p *RemoteSnapshotProvider) GetRemoteSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error) {
	client := p.clientFactory.CurrentClient()
	snapshotWithMetadata, err := client.GetSnapshot(ctx, rootName)
	if err != nil {
		return domain.Snapshot{}, err
	}

	return snapshotWithMetadata.Snapshot, nil
}
