package alias

import "insync/internal/domain"

type IAliasProvider interface {
	GetAliases() map[domain.NodeName]string
	SetAlias(newAlias string, nodeName domain.NodeName) error
	RemoveAlias(aliasName string) error
}
