package interfaces

import (
	"io"
	"io/fs"
)

type IFileSystem interface {
	ReadDir(fullPath string) ([]fs.DirEntry, error)
	WalkDir(fullPath string, walkFn func(path string, d fs.DirEntry, err error) error) error
	Open(fullPath string) (io.ReadCloser, error)
	Remove(fullPath string) error
	CreateTempFile(dir, pattern string) (io.WriteCloser, string, error)
	Rename(oldPath, newPath string) error
	MkdirAll(path string, perm fs.FileMode) error
	Create(fullPath string) (io.WriteCloser, error)
}
