package connect

import (
	"errors"
	"insync/internal/domain"
)

type ConnectUseCase struct {
	connectionManager IConnectionManager
}

var (
	ErrInvalidConnectUseCaseOptions = errors.New("Все поля ConnectUseCaseOptions должны быть заполнены")
)

func NewConnectUseCase(connectionManager IConnectionManager) (*ConnectUseCase, error) {
	if connectionManager == nil {
		return nil, ErrInvalidConnectUseCaseOptions
	}
	return &ConnectUseCase{
		connectionManager: connectionManager,
	}, nil
}

func (c *ConnectUseCase) CurrentNodeName() (domain.NodeName, error) {
	return c.connectionManager.CurrentNodeName()
}

func (c *ConnectUseCase) ConnectToNode(nodeName domain.NodeName) error {
	return c.connectionManager.ConnectToNode(nodeName)
}
