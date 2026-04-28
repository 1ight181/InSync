package hash

import "insync/internal/domain"

type IDirtyPathsRepository interface {
	GetDirtyPaths() (map[domain.ScopedPath]struct{}, error)
	SetDirtyPath(scopedPath domain.ScopedPath) error
	RemoveDirtyPath(scopedPath domain.ScopedPath) error
}
