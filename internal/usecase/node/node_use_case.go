package connect

import "insync/internal/domain"

type NodeUseCase struct {
	nodeNamesBrowser INodeNamesBrowser
}

type NodeUseCaseOptions struct {
	NodeNamesBrowser INodeNamesBrowser
}

func NewNodeUseCase(opts NodeUseCaseOptions) *NodeUseCase {
	if opts.NodeNamesBrowser == nil {
		panic("Все поля NodeUseCaseOptions должны быть заполнены")
	}
	return &NodeUseCase{
		nodeNamesBrowser: opts.NodeNamesBrowser,
	}
}

func (c *NodeUseCase) ShowNodeNames() (chan domain.NodeName, error) {
	return c.nodeNamesBrowser.BrowseNodeNames()
}
