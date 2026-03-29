package filemanager

import (
	"insync/internal/domain"
	"insync/internal/interfaces"
	"os"
)

type FileManager struct {
	pathResolver interfaces.IPathResolver
	hashResolver interfaces.IHashResolver
}

type FileManagerOptions struct {
	PathResolver interfaces.IPathResolver
	HashResolver interfaces.IHashResolver
}

func NewFileManager(opts FileManagerOptions) interfaces.IFileManager {
	if opts.PathResolver == nil || opts.HashResolver == nil {
		panic("Все поля FileManagerOptions должны быть заполнены")
	}
	return &FileManager{
		pathResolver: opts.PathResolver,
		hashResolver: opts.HashResolver,
	}
}

func (f *FileManager) GetFileList(rootName string, relativePath string) ([]domain.FileEntry, error) {
	resolvedPath, err := f.pathResolver.ResolvePath(rootName, relativePath)
	if err != nil {
		return nil, err
	}

	dirEntries, err := os.ReadDir(resolvedPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range dirEntries {
		entryInfo := entry.Info()


	return nil, nil
}