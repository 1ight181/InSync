package hash

import "insync/internal/domain"

type IHashCache interface {
	GetHashCache(fullPath domain.Path) (string, error)
	SetHashCache(fullPath domain.Path, hash string) error
}
