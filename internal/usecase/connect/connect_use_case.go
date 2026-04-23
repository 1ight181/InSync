package connect

import (
	"errors"
	"insync/internal/domain"
)

type ConnectUseCase struct {
	connectionManager IConnectionManager
}

type ConnectUseCaseOptions struct {
	ConnectionManager IConnectionManager
}

var (
	ErrInvalidConnectUseCaseOptions = errors.New("Все поля ConnectUseCaseOptions должны быть заполнены")
)

func NewConnectUseCase(opts ConnectUseCaseOptions) (*ConnectUseCase, error) {
	if opts.ConnectionManager == nil {
		return nil, ErrInvalidConnectUseCaseOptions
	}
	return &ConnectUseCase{
		connectionManager: opts.ConnectionManager,
	}, nil
}

func (c *ConnectUseCase) CurrentNodeName() (domain.NodeName, error) {
	return c.connectionManager.CurrentNodeName()
}

func (c *ConnectUseCase) ConnectToNode(nodeName domain.NodeName) error {
	return c.connectionManager.ConnectToNode(nodeName)
}
