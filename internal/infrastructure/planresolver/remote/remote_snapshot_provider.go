package remote

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type RemoteSnapshotProvider struct {
	clientFactory IClientFactory
}

type RemoteSnapshotProviderOptions struct {
	ClientFactory IClientFactory
}

var (
	ErrInvalidOpts = errors.New("Все поля RemoteSnapshotProviderOptions должны быть заполнены")
)

func NewRemoteSnapshotProvider(options RemoteSnapshotProviderOptions) (*RemoteSnapshotProvider, error) {
	if options.ClientFactory == nil {
		return nil, ErrInvalidOpts
	}
	return &RemoteSnapshotProvider{clientFactory: options.ClientFactory}, nil
}

func (p *RemoteSnapshotProvider) GetRemoteSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error) {
	client, err := p.clientFactory.CurrentClient()
	if err != nil {
		return domain.Snapshot{}, err
	}
	snapshot, err := client.GetSnapshot(ctx, rootName)
	if err != nil {
		return domain.Snapshot{}, err
	}

	return snapshot, nil
}
