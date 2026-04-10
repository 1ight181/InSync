package connection

import "insync/internal/domain"

type INodeNameResolver interface {
	ResolveToMDnsUrl(nodeName domain.NodeName) (string, error)
}
