package connect

import (
	"context"
	"errors"
	"insync/internal/domain"
)

type NodeUseCase struct {
	nodeNamesBrowser INodeNamesBrowser
}

type NodeUseCaseOptions struct {
	NodeNamesBrowser INodeNamesBrowser
}

var (
	ErrInvalidNodeUseCaseOptions = errors.New("Все поля NodeUseCaseOptions должны быть заполнены")
)

func NewNodeUseCase(opts NodeUseCaseOptions) (*NodeUseCase, error) {
	if opts.NodeNamesBrowser == nil {
		return nil, ErrInvalidNodeUseCaseOptions
	}
	return &NodeUseCase{
		nodeNamesBrowser: opts.NodeNamesBrowser,
	}, nil
}

func (c *NodeUseCase) ShowNodeNames(ctx context.Context) (chan domain.NodeName, error) {
	return c.nodeNamesBrowser.BrowseNodeNames(ctx)
}
