package filemanager

import (
	"insync/internal/domain"
	"io"
	"io/fs"
)

type IFileSystem interface {
	ReadDir(fullPath domain.Path) ([]fs.DirEntry, error)
	WalkDir(fullPath domain.Path, walkFn func(path string, d fs.DirEntry, err error) error) error
	Open(fullPath domain.Path) (fs.File, error)
	Remove(fullPath domain.Path) error
	CreateTempFile(dir domain.Path, pattern string) (io.ReadWriteCloser, domain.Path, error)
	Rename(oldPath, newPath domain.Path) error
	MkdirAll(path domain.Path, perm fs.FileMode) error
	AtomicWrite(fullPath domain.Path, data io.Reader) error
	Stat(fullPath domain.Path) (fs.FileInfo, error)
}
