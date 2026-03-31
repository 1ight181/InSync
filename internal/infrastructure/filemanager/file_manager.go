package filemanager

import (
	"context"
	"insync/internal/domain"
	"insync/internal/interfaces"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

type FileManager struct {
	pathResolver interfaces.IRootResolver
	hashResolver interfaces.IHashManager
	fileSystem   interfaces.IFileSystem

	logger    *slog.Logger
	loggerCtx context.Context
}

type FileManagerOptions struct {
	PathResolver interfaces.IRootResolver
	HashResolver interfaces.IHashManager
	FileSystem   interfaces.IFileSystem

	Logger    *slog.Logger
	LoggerCtx context.Context
}

func NewFileManager(opts FileManagerOptions) interfaces.IFileManager {
	if opts.PathResolver == nil || opts.HashResolver == nil || opts.FileSystem == nil || opts.Logger == nil || opts.LoggerCtx == nil {
		panic("Все поля FileManagerOptions должны быть заполнены")
	}
	return &FileManager{
		pathResolver: opts.PathResolver,
		hashResolver: opts.HashResolver,
		fileSystem:   opts.FileSystem,

		logger:    opts.Logger,
		loggerCtx: opts.LoggerCtx,
	}
}

func (f *FileManager) GetFileList(rootName string, relativePath string) ([]domain.FileEntry, error) {
	resolvedPath, err := f.pathResolver.ResolveRoot(rootName, relativePath)
	if err != nil {
		return nil, err
	}

	dirEntries, err := f.fileSystem.ReadDir(resolvedPath)
	if err != nil {
		return nil, err
	}

	var fileEntries []domain.FileEntry
	for _, entry := range dirEntries {
		entryInfo, err := entry.Info()
		if err != nil {
			return nil, err
		}

		fileMetadata := f.createMetadata(entryInfo)

		fullPath := filepath.Join(resolvedPath, entryInfo.Name())
		resourceContent := f.createResourceContent(entryInfo, fullPath)

		hash, err := f.hashResolver.ResolveHash(resourceContent, fileMetadata)
		if err != nil {
			return nil, err
		}

		fileInfo, err := domain.NewFileInfo(fileMetadata, hash)
		if err != nil {
			return nil, err
		}

		fileEntry, err := domain.NewFileEntry(rootName, relativePath, fileInfo)
		if err != nil {
			return nil, err
		}

		fileEntries = append(fileEntries, fileEntry)
	}

	return fileEntries, nil
}

func (f *FileManager) openContentDir(fullPath string) (io.ReadCloser, error) {
	pipeReader, pipeWriter, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	go func() {
		defer pipeWriter.Close()
		err := f.fileSystem.WalkDir(fullPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() {
				file, err := f.openContentFile(path)
				if err != nil {
					return err
				}
				defer file.Close()
				_, err = io.Copy(pipeWriter, file)
				if err != nil {
					return err
				}
				return nil
			}

			reader, err := f.openContentDir(path)
			if err != nil {
				return err
			}
			defer reader.Close()
			_, err = io.Copy(pipeWriter, reader)
			return err

		})
		if err != nil {
			f.logger.LogAttrs(
				f.loggerCtx,
				slog.LevelError,
				"Ошибка при чтении содержимого директории",
				slog.String("fullPath", fullPath),
				slog.String("error", err.Error()),
			)
		}
	}()

	return pipeReader, nil
}

func (f *FileManager) openContentFile(fullPath string) (io.ReadCloser, error) {
	return f.fileSystem.Open(fullPath)
}

func (f *FileManager) createMetadata(entryInfo fs.FileInfo) domain.FileMetadata {
	return domain.FileMetadata{
		ModifiedUnix: uint32(entryInfo.ModTime().Unix()),
		SizeBytes:    uint32(entryInfo.Size()),
		IsDirectory:  entryInfo.IsDir(),
	}
}

func (f *FileManager) createResourceContent(entryInfo fs.FileInfo, fullPath string) domain.ResourceContent {
	if entryInfo.IsDir() {
		return domain.ResourceContent{
			Path:        fullPath,
			OpenContent: f.openContentDir,
		}
	} else {
		return domain.ResourceContent{
			Path:        fullPath,
			OpenContent: f.openContentFile,
		}
	}
}
