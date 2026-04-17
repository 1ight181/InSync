package filesys

import (
	"insync/internal/domain"
	"io"
	"io/fs"
	"os"
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

func (f *FileSystem) CreateTempFile(dir domain.Path, pattern string) (io.ReadWriteCloser, domain.Path, error) {
	file, err := os.CreateTemp(dir.String(), pattern)
	fileFullPath, err := dir.Join(file.Name())
	if err != nil {
		return nil, "", err
	}
	return file, fileFullPath, err
}

func (f *FileSystem) Rename(oldPath, newPath domain.Path) error {
	return os.Rename(oldPath.String(), newPath.String())
}

func (f *FileSystem) MkdirAll(path domain.Path, perm fs.FileMode) error {
	return os.MkdirAll(path.String(), perm)
}

func (f *FileSystem) Create(fullPath domain.Path) (io.ReadWriteCloser, error) {
	return os.Create(fullPath.String())
}

func (f *FileSystem) Stat(fullPath domain.Path) (fs.FileInfo, error) {
	return os.Stat(fullPath.String())
}
