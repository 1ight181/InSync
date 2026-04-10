package connect

import "insync/internal/domain"

type IConnectionManager interface {
	ConnectToNode(nodeName domain.NodeName) error
	CurrentNodeName() (nodeName domain.NodeName)
}
