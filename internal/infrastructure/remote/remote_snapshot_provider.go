package remote

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type RemoteSnapshotProvider struct {
	clientFactory IClientFactory
}

var (
	ErrInvalidOpts = errors.New("Все поля RemoteSnapshotProvider не должны быть nil")
)

func NewRemoteSnapshotProvider(clientFactory IClientFactory) (*RemoteSnapshotProvider, error) {
	if clientFactory == nil {
		return nil, ErrInvalidOpts
	}
	return &RemoteSnapshotProvider{clientFactory: clientFactory}, nil
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
