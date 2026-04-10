package filesys

import (
	"context"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

type FileSystem struct {
	logger    *slog.Logger
	loggerCtx context.Context
}

type FileSystemOptions struct {
	Logger *slog.Logger
}

func NewFileSystem(opts FileSystemOptions) *FileSystem {
	if opts.Logger == nil {
		panic("Не все обязательные параметры были переданы при инициализации FileSystem")
	}
	loggerCtx := context.Background()
	return &FileSystem{
		logger:    opts.Logger,
		loggerCtx: loggerCtx,
	}
}

func (f *FileSystem) ReadDir(fullPath string) ([]fs.DirEntry, error) {
	return os.ReadDir(fullPath)
}

func (f *FileSystem) WalkDir(fullPath string, walkFn func(path string, d fs.DirEntry, err error) error) error {
	return fs.WalkDir(os.DirFS("/"), fullPath, walkFn)
}

func (f *FileSystem) Open(fullPath string) (fs.File, error) {
	return os.Open(fullPath)
}

func (f *FileSystem) Remove(fullPath string) error {
	return os.Remove(fullPath)
}

func (f *FileSystem) CreateTempFile(dir, pattern string) (io.ReadWriteCloser, string, error) {
	file, err := os.CreateTemp(dir, pattern)
	fileFullPath := filepath.Join(dir, file.Name())
	return file, fileFullPath, err
}

func (f *FileSystem) Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

func (f *FileSystem) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (f *FileSystem) Create(fullPath string) (io.ReadWriteCloser, error) {
	return os.Create(fullPath)
}
