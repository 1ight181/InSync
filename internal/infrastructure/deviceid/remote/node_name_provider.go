package deviceid

import "insync/internal/domain"

type INodeNameProvider interface {
	CurrentNodeName() (domain.NodeName, error)
}
