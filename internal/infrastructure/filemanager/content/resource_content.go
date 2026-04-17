package domain

import (
	"insync/internal/domain"
	"io"
)

type ResourceContent struct {
	FullPath     domain.Path
	RelativePath domain.Path

	OpenContent func(fullPath domain.Path, relativePath domain.Path) (io.ReadCloser, error)
}
