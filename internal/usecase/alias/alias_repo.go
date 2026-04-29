package alias

import "insync/internal/domain"

type IAliasRepository interface {
	GetAliases() ([]string, error)
	SetAlias(newAlias string, nodeName domain.NodeName) error
	RemoveAlias(aliasName string) error
}
