package alias

import (
	"context"
	"insync/internal/domain"
)

type IAliasRepository interface {
	GetAliases(ctx context.Context) (map[domain.NodeName]string, error)
	SetAlias(ctx context.Context, newAlias string, nodeName domain.NodeName) error
	RemoveAlias(ctx context.Context, aliasName string) error
}
