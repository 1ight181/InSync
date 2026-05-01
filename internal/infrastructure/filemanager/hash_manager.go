package filemanager

import (
	"context"
	"insync/internal/domain"
	cont "insync/internal/infrastructure/filemanager/content"
)

type IHashManager interface {
	ResolveHash(ctx context.Context, resourceContent cont.ResourceContent, rootName domain.RootName) (string, error)
	MarkDirty(ctx context.Context, scopedPath domain.ScopedPath) error
	ResolveWithForceRecalc(ctx context.Context, resourceContent cont.ResourceContent, rootName domain.RootName) (string, error)
}
