package connect

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type NodeUseCase struct {
	nodeNamesBrowser      INodeNamesBrowser
	localDeviceIdResolver ILocalDeviceIdResolver
	aliasProvider         IAliasProvider
}

type NodeUseCaseOptions struct {
	NodeNamesBrowser      INodeNamesBrowser
	LocalDeviceIdResolver ILocalDeviceIdResolver
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

func (c *NodeUseCase) ShowNodeNames(ctx context.Context) (chan domain.NodeName, error) {
	rawNodeNamesChan, err := c.nodeNamesBrowser.BrowseNodeNames(ctx)
	if err != nil {
		return nil, err
	}

	localDeviceId, err := c.localDeviceIdResolver.Resolve()
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

			validNodeName := rawNodeName

			alias := c.aliasProvider.GetAlias(rawNodeName)
			if alias != "" {
				validNodeName = domain.NodeName(alias)
			}

			validNodeNamesChan <- validNodeName
		}
	}()

	return validNodeNamesChan, nil
}
