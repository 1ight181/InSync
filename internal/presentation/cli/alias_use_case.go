package cli

import "insync/internal/domain"

type IAliasUseCase interface {
	SetAlias(newAlias string, nodeName domain.NodeName) error
	RemoveAlias(aliasName string) error
	GetAliases() ([]string, error)
}
