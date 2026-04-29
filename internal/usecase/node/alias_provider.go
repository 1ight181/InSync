package connect

import "insync/internal/domain"

type IAliasProvider interface {
	GetAlias(nodeName domain.NodeName) string
}
