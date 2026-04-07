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

	tempDir string

	logger    *slog.Logger
	loggerCtx context.Context
}

type FileManagerOptions struct {
	PathResolver interfaces.IRootResolver
	HashResolver interfaces.IHashManager
	FileSystem   interfaces.IFileSystem

	TempDir string

	Logger    *slog.Logger
	LoggerCtx context.Context
}

func NewFileManager(opts FileManagerOptions) interfaces.IFileManager {
	if opts.PathResolver == nil ||
		opts.HashResolver == nil ||
		opts.FileSystem == nil ||
		opts.TempDir == "" ||
		opts.Logger == nil ||
		opts.LoggerCtx == nil {
		panic("Все поля FileManagerOptions должны быть заполнены")
	}
	return &FileManager{
		pathResolver: opts.PathResolver,
		hashResolver: opts.HashResolver,
		fileSystem:   opts.FileSystem,

		tempDir: opts.TempDir,

		logger:    opts.Logger,
		loggerCtx: opts.LoggerCtx,
	}
}
func (f *FileManager) GetFileList(rootName string) ([]domain.FileEntry, error) {
	resolvedRootPath, err := f.pathResolver.ResolveRoot(rootName, "")
	if err != nil {
		return nil, err
	}

	return f.collectFileEntriesRecursive(rootName, resolvedRootPath, "")
}

func (f *FileManager) RenameFile(rootName string, oldRelativePath string, newRelativePath string) error {
	oldResolvedPath, err := f.pathResolver.ResolveRoot(rootName, oldRelativePath)
	if err != nil {
		return err
	}

	newResolvedPath, err := f.pathResolver.ResolveRoot(rootName, newRelativePath)
	if err != nil {
		return err
	}

	if err := f.fileSystem.Rename(oldResolvedPath, newResolvedPath); err != nil {
		return err
	}

	return f.hashResolver.MarkDirty(newResolvedPath)
}

func (f *FileManager) DeleteFile(rootName string, relativePath string) error {
	resolvedPath, err := f.pathResolver.ResolveRoot(rootName, relativePath)
	if err != nil {
		return err
	}

	if err := f.fileSystem.Remove(resolvedPath); err != nil {
		return err
	}

	return f.hashResolver.MarkDirty(resolvedPath)
}

func (f *FileManager) GetFile(rootName string, relativePath string) (io.ReadCloser, error) {
	resolvedPath, err := f.pathResolver.ResolveRoot(rootName, relativePath)
	if err != nil {
		return nil, err
	}

	return f.openContentFile(resolvedPath)
}

func (f *FileManager) PutFile(rootName string, relativePath string, content io.Reader) error {
	resolvedPath, err := f.pathResolver.ResolveRoot(rootName, relativePath)
	if err != nil {
		return err
	}

	tempFilePath, err := f.writeContentToTemp(content)
	if err != nil {
		return err
	}

	if err := f.moveTempToDestination(tempFilePath, resolvedPath); err != nil {
		return err
	}

	return f.hashResolver.MarkDirty(resolvedPath)
}

func (f *FileManager) writeContentToTemp(content io.Reader) (string, error) {
	tempFile, tempFilePath, err := f.fileSystem.CreateTempFile(f.tempDir, "filemanager_temp_*")
	if err != nil {
		return "", err
	}

	defer tempFile.Close()

	_, err = io.Copy(tempFile, content)
	if err != nil {
		if err := f.fileSystem.Remove(tempFilePath); err != nil {
			f.logger.LogAttrs(
				f.loggerCtx,
				slog.LevelError,
				"Не удалось удалить временный файл при ошибке записи в методе writeContentToTemp",
				slog.String("tempFilePath", tempFilePath),
			)
		}
		return "", err
	}

	return tempFilePath, nil

}

func (f *FileManager) moveTempToDestination(tempFilePath string, destPath string) error {
	defer func() {
		if err := f.fileSystem.Remove(tempFilePath); err != nil {
			f.logger.LogAttrs(
				f.loggerCtx,
				slog.LevelError,
				"Не удалось удалить временный файл при ошибке перемещения в методе moveTempToDestination",
				slog.String("tempFilePath", tempFilePath),
			)
		}
	}()

	if err := f.fileSystem.MkdirAll(filepath.Dir(destPath), 0755); err != nil {

		return err
	}

	if err := f.fileSystem.Rename(tempFilePath, destPath); err != nil {
		// fallback при ошибке перемещения между fs
		srcFile, err := f.fileSystem.Open(tempFilePath)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := f.fileSystem.Create(destPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return err
		}
	}

	return nil
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

			if d.Type()&os.ModeSymlink != 0 {
				f.logger.LogAttrs(
					f.loggerCtx,
					slog.LevelWarn,
					"Обнаружена и пропущена символическая ссылка при чтении директории",
					slog.String("fullPath", fullPath),
					slog.String("symlinkPath", path),
				)
				return nil
			}

			if d.IsDir() {
				return nil
			}

			file, err := f.openContentFile(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(pipeWriter, file)
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

func (f *FileManager) collectFileEntriesRecursive(
	rootName string,
	currentAbsolutePath string,
	currentRelativePath string,
) ([]domain.FileEntry, error) {

	directoryEntries, err := f.fileSystem.ReadDir(currentAbsolutePath)
	if err != nil {
		return nil, err
	}

	var collectedEntries []domain.FileEntry

	for _, directoryEntry := range directoryEntries {
		entryName := directoryEntry.Name()

		nextAbsolutePath := filepath.Join(currentAbsolutePath, entryName)
		nextRelativePath := filepath.Join(currentRelativePath, entryName)

		if directoryEntry.IsDir() {
			nestedEntries, err := f.collectFileEntriesRecursive(
				rootName,
				nextAbsolutePath,
				nextRelativePath,
			)
			if err != nil {
				return nil, err
			}

			collectedEntries = append(collectedEntries, nestedEntries...)
			continue
		}
		entryInfo, err := directoryEntry.Info()
		if err != nil {
			return nil, err
		}

		fileMetadata := f.createMetadata(entryInfo)
		resourceContent := f.createResourceContent(entryInfo, nextAbsolutePath)

		hashValue, err := f.hashResolver.ResolveHash(resourceContent, fileMetadata)
		if err != nil {
			return nil, err
		}

		fileInfo, err := domain.NewFileInfo(fileMetadata, hashValue)
		if err != nil {
			return nil, err
		}

		fileEntry, err := domain.NewFileEntry(rootName, nextRelativePath, fileInfo)
		if err != nil {
			return nil, err
		}

		collectedEntries = append(collectedEntries, fileEntry)
	}

	return collectedEntries, nil
}
