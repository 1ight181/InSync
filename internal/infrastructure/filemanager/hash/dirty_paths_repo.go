package hash

import (
	"context"
	"insync/internal/domain"
)

type IDirtyPathsRepository interface {
	GetDirtyPaths(ctx context.Context) (map[domain.ScopedPath]struct{}, error)
	SetDirtyPath(ctx context.Context, scopedPath domain.ScopedPath) error
	RemoveDirtyPath(ctx context.Context, scopedPath domain.ScopedPath) error
}
