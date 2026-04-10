package connect

import "insync/internal/domain"

type ConnectUseCase struct {
	nodeNamesBrowser  INodeNamesBrowser
	connectionManager IConnectionManager
}

type ConnectUseCaseOptions struct {
	NodeNamesBrowser  INodeNamesBrowser
	ConnectionManager IConnectionManager
}

func NewConnectUseCase(opts ConnectUseCaseOptions) *ConnectUseCase {
	if opts.NodeNamesBrowser == nil ||
		opts.ConnectionManager == nil {
		panic("Все поля ConnectUseCaseOptions должны быть заполнены")
	}
	return &ConnectUseCase{
		nodeNamesBrowser:  opts.NodeNamesBrowser,
		connectionManager: opts.ConnectionManager,
	}
}

func (c *ConnectUseCase) ShowNodes() (chan domain.NodeName, error) {
	return c.nodeNamesBrowser.BrowseNodeNames()
}

func (c *ConnectUseCase) CurrentNode() domain.NodeName {
	return c.connectionManager.CurrentNodeName()
}

func (c *ConnectUseCase) ConnectToNode(nodeName domain.NodeName) error {
	return c.connectionManager.ConnectToNode(nodeName)
}
