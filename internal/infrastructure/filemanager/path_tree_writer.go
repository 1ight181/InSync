package filemanager

import "insync/internal/domain"

type IPathTreeWriter interface {
	AddPath(fullPath domain.Path) error
	RemovePath(fullPath domain.Path) error
}
