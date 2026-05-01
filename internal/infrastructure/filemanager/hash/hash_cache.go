package hash

import (
	"context"
	"insync/internal/domain"
)

type IHashCache interface {
	GetHashCache(fullPath domain.Path) (string, error)
	SetHashCache(ctx context.Context, fullPath domain.Path, hash string) error
}
