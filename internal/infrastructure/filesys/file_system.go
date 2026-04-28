package filesys

import (
	"insync/internal/domain"
	"io"
	"io/fs"
	"os"

	atom "github.com/natefinch/atomic"
)

type FileSystem struct{}

func NewFileSystem() *FileSystem {
	return &FileSystem{}
}

func (f *FileSystem) ReadDir(fullPath domain.Path) ([]fs.DirEntry, error) {
	return os.ReadDir(fullPath.String())
}

func (f *FileSystem) WalkDir(fullPath domain.Path, walkFn func(path string, d fs.DirEntry, err error) error) error {
	return fs.WalkDir(os.DirFS("/"), fullPath.String(), walkFn)
}

func (f *FileSystem) Open(fullPath domain.Path) (fs.File, error) {
	return os.Open(fullPath.String())
}

func (f *FileSystem) Remove(fullPath domain.Path) error {
	return os.Remove(fullPath.String())
}

func (f *FileSystem) Rename(oldPath, newPath domain.Path) error {
	return atom.ReplaceFile(oldPath.String(), newPath.String())
}

func (f *FileSystem) MkdirAll(path domain.Path, perm fs.FileMode) error {
	return os.MkdirAll(path.String(), perm)
}

func (f *FileSystem) Stat(fullPath domain.Path) (fs.FileInfo, error) {
	return os.Stat(fullPath.String())
}

func (f *FileSystem) AtomicWrite(fullPath domain.Path, data io.Reader) error {
	return atom.WriteFile(fullPath.String(), data)
}

func (f *FileSystem) Mkdir(fullPath domain.Path, perm fs.FileMode) error {
	return os.Mkdir(fullPath.String(), perm)
}
