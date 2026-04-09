package interfaces

import "insync/internal/domain"

type IConnectUseCase interface {
	ConnectToNode(nodeName string) error
	CurrentNode() (string, error)
	GetAllNodes() ([]domain.Node, error)
}
