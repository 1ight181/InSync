package alias

import "insync/internal/domain"

type IAliasRepository interface {
	GetAliases() (map[domain.NodeName]string, error)
	SetAlias(newAlias string, nodeName domain.NodeName) error
	RemoveAlias(aliasName string) error
}
