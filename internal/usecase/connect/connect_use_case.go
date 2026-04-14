package connect

import "insync/internal/domain"

type ConnectUseCase struct {
	connectionManager IConnectionManager
}

type ConnectUseCaseOptions struct {
	ConnectionManager IConnectionManager
}

func NewConnectUseCase(opts ConnectUseCaseOptions) *ConnectUseCase {
	if opts.ConnectionManager == nil {
		panic("Все поля ConnectUseCaseOptions должны быть заполнены")
	}
	return &ConnectUseCase{
		connectionManager: opts.ConnectionManager,
	}
}

func (c *ConnectUseCase) CurrentNodeName() domain.NodeName {
	return c.connectionManager.CurrentNodeName()
}

func (c *ConnectUseCase) ConnectToNode(nodeName domain.NodeName) error {
	return c.connectionManager.ConnectToNode(nodeName)
}
