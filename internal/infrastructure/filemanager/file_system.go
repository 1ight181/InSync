package filemanager

import (
	"io"
	"io/fs"
)

type IFileSystem interface {
	ReadDir(fullPath string) ([]fs.DirEntry, error)
	WalkDir(fullPath string, walkFn func(path string, d fs.DirEntry, err error) error) error
	Open(fullPath string) (fs.File, error)
	Remove(fullPath string) error
	CreateTempFile(dir, pattern string) (io.ReadWriteCloser, string, error)
	Rename(oldPath, newPath string) error
	MkdirAll(path string, perm fs.FileMode) error
	Create(fullPath string) (io.ReadWriteCloser, error)
}
