package cli

import (
	"context"
	"insync/internal/domain"
)

type IRootUseCase interface {
	AddRoot(ctx context.Context, rootName domain.RootName, relativePath domain.Path) error
	RemoveRoot(ctx context.Context, rootName domain.RootName) error
	GetRoots() (map[domain.RootName]domain.Path, error)
}
