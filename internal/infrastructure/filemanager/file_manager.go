package filemanager

import (
	"context"
	"insync/internal/domain"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

type FileManager struct {
	rootResolver   IRootResolver
	hashManager    IHashManager
	fileSystem     IFileSystem
	pathTreeWriter IPathTreeWriter

	tempDir string

	logger    *slog.Logger
	loggerCtx context.Context
}

type FileManagerOptions struct {
	RootResolver   IRootResolver
	HashManager    IHashManager
	FileSystem     IFileSystem
	PathTreeWriter IPathTreeWriter

	TempDir string

	Logger    *slog.Logger
	LoggerCtx context.Context
}

func NewFileManager(opts FileManagerOptions) *FileManager {
	if opts.RootResolver == nil ||
		opts.HashManager == nil ||
		opts.FileSystem == nil ||
		opts.PathTreeWriter == nil ||
		opts.TempDir == "" ||
		opts.Logger == nil ||
		opts.LoggerCtx == nil {
		panic("Все поля FileManagerOptions должны быть заполнены")
	}
	return &FileManager{
		rootResolver:   opts.RootResolver,
		hashManager:    opts.HashManager,
		fileSystem:     opts.FileSystem,
		pathTreeWriter: opts.PathTreeWriter,

		tempDir: opts.TempDir,

		logger:    opts.Logger,
		loggerCtx: opts.LoggerCtx,
	}
}
func (f *FileManager) GetFileList(ctx context.Context, rootName domain.RootName) ([]domain.FileEntry, error) {
	root := rootName.String()
	resolvedRootPath, err := f.rootResolver.ResolveRoot(root, "")
	if err != nil {
		return nil, err
	}

	return f.collectFileEntriesRecursive(ctx, root, resolvedRootPath, "")
}

func (f *FileManager) RenameFile(ctx context.Context, rootName domain.RootName, oldRelativePath domain.RelativePath, newRelativePath domain.RelativePath) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	root := rootName.String()

	oldRelPath := oldRelativePath.String()
	newRelPath := newRelativePath.String()

	oldResolvedPath, err := f.rootResolver.ResolveRoot(root, oldRelPath)
	if err != nil {
		return err
	}

	newResolvedPath, err := f.rootResolver.ResolveRoot(root, newRelPath)
	if err != nil {
		return err
	}

	if err := f.fileSystem.Rename(oldResolvedPath, newResolvedPath); err != nil {
		return err
	}

	if err := f.pathTreeWriter.RemovePath(oldResolvedPath); err != nil {
		return err
	}

	if err := f.pathTreeWriter.AddPath(newResolvedPath); err != nil {
		return err
	}

	return f.hashManager.MarkDirty(newResolvedPath)
}

func (f *FileManager) DeleteFile(ctx context.Context, rootName domain.RootName, relativePath domain.RelativePath) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	root := rootName.String()
	relPath := relativePath.String()

	resolvedPath, err := f.rootResolver.ResolveRoot(root, relPath)
	if err != nil {
		return err
	}

	if err := f.fileSystem.Remove(resolvedPath); err != nil {
		return err
	}

	if err := f.pathTreeWriter.RemovePath(resolvedPath); err != nil {
		return err
	}

	return f.hashManager.MarkDirty(resolvedPath)
}

func (f *FileManager) GetFile(ctx context.Context, rootName domain.RootName, relativePath domain.RelativePath) (io.ReadCloser, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	root := rootName.String()
	relPath := relativePath.String()

	resolvedPath, err := f.rootResolver.ResolveRoot(root, relPath)
	if err != nil {
		return nil, err
	}

	return f.openContentFile(resolvedPath)
}

func (f *FileManager) PutFile(ctx context.Context, rootName domain.RootName, relativePath domain.RelativePath, content io.Reader) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	root := rootName.String()
	relPath := relativePath.String()

	resolvedPath, err := f.rootResolver.ResolveRoot(root, relPath)
	if err != nil {
		return err
	}

	tempFilePath, err := f.writeContentToTemp(content)
	if err != nil {
		return err
	}

	if err := f.moveTempToDestination(ctx, tempFilePath, resolvedPath); err != nil {
		return err
	}

	if err := f.pathTreeWriter.AddPath(resolvedPath); err != nil {
		return err
	}

	return f.hashManager.MarkDirty(resolvedPath)
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

func (f *FileManager) moveTempToDestination(ctx context.Context, tempFilePath string, destPath string) error {
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
		if ctx.Err() != nil {
			return ctx.Err()
		}
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
	ctx context.Context,
	rootName string,
	currentAbsolutePath string,
	currentRelativePath string,
) ([]domain.FileEntry, error) {

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

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
				ctx,
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

		hashValue, err := f.hashManager.ResolveHash(resourceContent, fileMetadata)
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
