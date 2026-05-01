package root

import (
	"context"
	"insync/internal/domain"
)

type IRootRegistrar interface {
	AddRoot(ctx context.Context, rootName domain.RootName, rootPath domain.Path) error
	RemoveRoot(ctx context.Context, rootName domain.RootName) error
	GetRoots() (map[domain.RootName]domain.Path, error)
}
