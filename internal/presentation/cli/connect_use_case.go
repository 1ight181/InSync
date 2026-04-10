package cli

import "insync/internal/domain"

type IConnectUseCase interface {
	ConnectToNode(nodeName domain.NodeName) error
	CurrentNodeName() (nodeName domain.NodeName)
	ShowNodeNames() (nodeNamesChan chan domain.NodeName, err error)
}
