package connect

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type NodeUseCase struct {
	nodeNamesBrowser      INodeNamesBrowser
	localDeviceIdResolver ILocalDeviceIdProvider
	aliasProvider         IAliasProvider
}

type NodeUseCaseOptions struct {
	NodeNamesBrowser      INodeNamesBrowser
	LocalDeviceIdResolver ILocalDeviceIdProvider
	AliasProvider         IAliasProvider
}

var (
	ErrInvalidNodeUseCaseOptions = errors.New("Все поля NodeUseCaseOptions должны быть заполнены")
)

func NewNodeUseCase(opts NodeUseCaseOptions) (*NodeUseCase, error) {
	if opts.NodeNamesBrowser == nil ||
		opts.LocalDeviceIdResolver == nil ||
		opts.AliasProvider == nil {
		return nil, ErrInvalidNodeUseCaseOptions
	}
	return &NodeUseCase{
		nodeNamesBrowser:      opts.NodeNamesBrowser,
		localDeviceIdResolver: opts.LocalDeviceIdResolver,
		aliasProvider:         opts.AliasProvider,
	}, nil
}

// Возвращает канал с мапой имя узла - алиас
func (c *NodeUseCase) ShowNodeNames(ctx context.Context) (chan domain.NodeNameWithAlias, error) {
	rawNodeNamesChan, err := c.nodeNamesBrowser.BrowseNodeNames(ctx)
	if err != nil {
		return nil, err
	}

	localDeviceId, err := c.localDeviceIdResolver.GetCurrentLocalDeviceId()
	if err != nil {
		return nil, err
	}

	validNodeNamesChan := make(chan domain.NodeNameWithAlias, 10)
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

			validNodeName := rawNodeName

			alias := c.aliasProvider.GetAlias(rawNodeName)
			validNodeNamesChan <- domain.NodeNameWithAlias{
				NodeName: validNodeName,
				Alias:    alias,
			}
		}
	}()

	return validNodeNamesChan, nil
}
