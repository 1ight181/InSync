package cache

import (
	"context"
	"insync/internal/domain"
)

type IHashCacheRepository interface {
	GetHashCache(ctx context.Context) (map[domain.Path]string, error)
	SetHashCache(ctx context.Context, fullPath domain.Path, hash string) error
}
