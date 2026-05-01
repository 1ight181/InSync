package cli

import "insync/internal/domain"

type IAliasUseCase interface {
	SetAlias(newAlias string, nodeName domain.NodeName) error
	RemoveAlias(aliasName string) error
	GetNodeByAlias(aliasName string) (domain.NodeName, bool)
	GetAliasByNode(nodeName domain.NodeName) (string, bool)
	GetAliases() map[domain.NodeName]string
}
