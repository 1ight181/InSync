package creator

import (
	"insync/internal/domain"
	"io"
)

type IFileSystem interface {
	AtomicWrite(fullPath domain.Path, data io.Reader) error
}
