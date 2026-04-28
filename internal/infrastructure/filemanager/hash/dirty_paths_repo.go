package hash

import "insync/internal/domain"

type IDirtyPathsRepository interface {
	GetDirtyPaths() (map[domain.ScopedPath]struct{}, error)
}
