package cli

import (
	"context"
	"insync/internal/domain"
)

type IAliasUseCase interface {
	SetAlias(ctx context.Context, newAlias string, nodeName domain.NodeName) error
	RemoveAlias(ctx context.Context, aliasName string) error
	GetNodeByAlias(aliasName string) (domain.NodeName, bool)
	GetAliasByNode(nodeName domain.NodeName) (string, bool)
	GetAliases() map[domain.NodeName]string
}
