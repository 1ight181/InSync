package cache

import "insync/internal/domain"

type IHashCacheLoader interface {
	GetHashCache() (map[domain.Path]string, error)
}
