package root

import (
	"context"
	"insync/internal/domain"
)

type IRootResolverRepository interface {
	GetRoots() (map[domain.RootName]domain.Path, error)
	AddRoot(ctx context.Context, rootName domain.RootName, rootPath domain.Path) error
	RemoveRoot(ctx context.Context, rootName domain.RootName) error
}
