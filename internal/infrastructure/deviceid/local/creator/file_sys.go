package creator

import (
	"insync/internal/domain"
	"io"
)

type IFileSystem interface {
	Create(fullPath domain.Path) (io.ReadWriteCloser, error)
}
