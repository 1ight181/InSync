package cache

import "insync/internal/domain"

type IHashCacheRepository interface {
	GetHashCache() (map[domain.Path]string, error)
	SetHashCache(fullPath domain.Path, hash string) error
}
