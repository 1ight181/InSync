package creator

import (
	"insync/internal/domain"
	"io"
	"io/fs"
)

type IFileSystem interface {
	AtomicWrite(fullPath domain.Path, data io.Reader) error
	Open(fullPath domain.Path) (fs.File, error)
}
