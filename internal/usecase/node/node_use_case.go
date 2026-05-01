package connect

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type NodeUseCase struct {
	nodeNamesBrowser      INodeNamesBrowser
	localDeviceIdResolver ILocalDeviceIdProvider
}

type NodeUseCaseOptions struct {
	NodeNamesBrowser      INodeNamesBrowser
	LocalDeviceIdResolver ILocalDeviceIdProvider
}

var (
	ErrInvalidNodeUseCaseOptions = errors.New("Все поля NodeUseCaseOptions должны быть заполнены")
)

func NewNodeUseCase(opts NodeUseCaseOptions) (*NodeUseCase, error) {
	if opts.NodeNamesBrowser == nil ||
		opts.LocalDeviceIdResolver == nil {
		return nil, ErrInvalidNodeUseCaseOptions
	}
	return &NodeUseCase{
		nodeNamesBrowser:      opts.NodeNamesBrowser,
		localDeviceIdResolver: opts.LocalDeviceIdResolver,
	}, nil
}

// Возвращает канал с мапой имя узла - алиас
func (c *NodeUseCase) ShowNodeNames(ctx context.Context) (chan domain.NodeName, error) {
	rawNodeNamesChan, err := c.nodeNamesBrowser.BrowseNodeNames(ctx)
	if err != nil {
		return nil, err
	}

	localDeviceId, err := c.localDeviceIdResolver.GetCurrentLocalDeviceId()
	if err != nil {
		return nil, err
	}

	validNodeNamesChan := make(chan domain.NodeName, 10)
	selfNodeName, err := domain.NewNodeName(localDeviceId.String())
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(validNodeNamesChan)
		for rawNodeName := range rawNodeNamesChan {
			if rawNodeName == selfNodeName {
				continue
			}

			validNodeNamesChan <- rawNodeName
		}
	}()

	return validNodeNamesChan, nil
}
