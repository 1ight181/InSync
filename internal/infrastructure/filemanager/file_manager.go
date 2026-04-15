package filemanager

import (
	"bytes"
	"context"
	"errors"
	"insync/internal/domain"
	cont "insync/internal/infrastructure/filemanager/content"
	"sort"
	"syscall"
	"time"

	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"go.uber.org/multierr"
)

type multiCloser struct {
	closers []io.Closer
}

func (m multiCloser) Close() error {
	var closeErr error
	for _, closer := range m.closers {
		if err := closer.Close(); err != nil {
			closeErr = multierr.Append(closeErr, err)
		}
	}
	return closeErr
}

func NewMultiCloser(closers ...io.Closer) io.Closer {
	return multiCloser{closers: closers}
}

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

func (f *FileManager) GetSnapshot(ctx context.Context, rootName domain.RootName) (domain.Snapshot, error) {
	resolvedRootPath, err := f.rootResolver.ResolveRoot(rootName, "")
	if err != nil {
		return domain.Snapshot{}, err
	}

	allEntries, err := f.collectAllFileEntries(ctx, resolvedRootPath)
	if err != nil {
		return domain.Snapshot{}, err
	}

	snapshot := domain.NewSnapshot(uint64(time.Now().Unix()), allEntries)

	return snapshot, nil
}

func (f *FileManager) RenameFile(ctx context.Context, rootName domain.RootName, oldPath domain.Path, newPath domain.Path) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	oldResolvedPath, err := f.rootResolver.ResolveRoot(rootName, oldPath)
	if err != nil {
		return err
	}

	newResolvedPath, err := f.rootResolver.ResolveRoot(rootName, newPath)
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

func (f *FileManager) DeleteFile(ctx context.Context, rootName domain.RootName, relativePath domain.Path) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	resolvedPath, err := f.rootResolver.ResolveRoot(rootName, relativePath)
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

func (f *FileManager) GetFile(ctx context.Context, rootName domain.RootName, relativePath domain.Path) (io.ReadCloser, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	resolvedFullPath, err := f.rootResolver.ResolveRoot(rootName, relativePath)
	if err != nil {
		return nil, err
	}

	return f.openFileContentWithHeader(resolvedFullPath, relativePath.String())
}

func (f *FileManager) PutFile(ctx context.Context, rootName domain.RootName, relativePath domain.Path, content io.Reader) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	resolvedPath, err := f.rootResolver.ResolveRoot(rootName, relativePath)
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

	if err := f.fileSystem.Rename(tempFilePath, destPath); errors.Is(err, syscall.EXDEV) {
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

func (f *FileManager) openDirContent(fullPath, relativePath string) (io.ReadCloser, error) {

	entries, err := f.fileSystem.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var readers []io.Reader
	var closers []io.Closer

	dirHeaderContent := bytes.NewReader([]byte(relativePath))
	readers = append(readers, dirHeaderContent)

	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 {
			f.logger.LogAttrs(
				f.loggerCtx,
				slog.LevelWarn,
				"Обнаружена и пропущена символическая ссылка при чтении директории",
				slog.String("fullPath", fullPath),
				slog.String("symlinkName", filepath.Join(fullPath, entry.Name())),
			)
			continue
		}

		if entry.IsDir() {
			subDirFullPath := filepath.Join(fullPath, entry.Name())
			subDirPath := filepath.Join(relativePath, entry.Name())
			subDirContent, err := f.openDirContent(subDirFullPath, subDirPath)
			if err != nil {
				f.logger.LogAttrs(
					f.loggerCtx,
					slog.LevelError,
					"Ошибка при чтении директории",
					slog.String("fullPath", subDirFullPath),
				)

				NewMultiCloser(closers...).Close()

				return nil, err
			}

			readers = append(readers, subDirContent)
			closers = append(closers, subDirContent)
			continue
		}

		fileFullPath := filepath.Join(fullPath, entry.Name())
		filePath := filepath.Join(relativePath, entry.Name())
		fileContent, err := f.openFileContentWithHeader(fileFullPath, filePath)
		if err != nil {
			f.logger.LogAttrs(
				f.loggerCtx,
				slog.LevelError,
				"Ошибка при чтении файла",
				slog.String("fullPath", fileFullPath),
			)

			NewMultiCloser(closers...).Close()

			return nil, err
		}

		readers = append(readers, fileContent)
		closers = append(closers, fileContent)
	}

	content := io.MultiReader(readers...)

	return struct {
		io.Reader
		io.Closer
	}{
		Reader: content,
		Closer: NewMultiCloser(closers...),
	}, nil

}

func (f *FileManager) openFileContentWithHeader(fullPath, relativePath string) (io.ReadCloser, error) {
	fileContent, err := f.fileSystem.Open(fullPath)
	if err != nil {
		return nil, err
	}

	headerContent := bytes.NewReader([]byte(relativePath))
	content := io.MultiReader(headerContent, fileContent)

	return struct {
		io.Reader
		io.Closer
	}{
		Reader: content,
		Closer: fileContent,
	}, nil
}

func (f *FileManager) openFileContent(fullPath, _ string) (io.ReadCloser, error) {
	return f.fileSystem.Open(fullPath)
}

func (f *FileManager) createMetadata(entryInfo fs.FileInfo) domain.FileMetadata {
	return domain.FileMetadata{
		ModifiedUnix: uint64(entryInfo.ModTime().Unix()),
		SizeBytes:    uint64(entryInfo.Size()),
		IsDirectory:  entryInfo.IsDir(),
	}
}

func (f *FileManager) collectAllFileEntries(
	ctx context.Context,
	rootAbsolutePath string,
) ([]domain.FileEntry, error) {

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	children, childrenSubtreeSize, err := f.collectFileEntriesRecursive(ctx, rootAbsolutePath, "")
	if err != nil {
		return nil, err
	}

	rootEntrySubtreeSize := uint64(1 + childrenSubtreeSize) // размер корневой директории

	rootEntry, err := f.createFileEntryForDirectory(ctx, rootAbsolutePath, "", rootEntrySubtreeSize)
	if err != nil {
		return nil, err
	}

	result := make([]domain.FileEntry, 0, len(children)+1)
	result = append(result, rootEntry)
	result = append(result, children...)

	return result, nil
}

func (f *FileManager) collectFileEntriesRecursive(
	ctx context.Context,
	currentAbsolutePath string,
	currentPath string,
) ([]domain.FileEntry, uint64, error) { // добавили возврат размера поддерева
	directoryEntries, err := f.fileSystem.ReadDir(currentAbsolutePath)
	if err != nil {
		return nil, 0, err
	}

	var collectedEntries []domain.FileEntry
	var subtreeSize uint64 = 0 // размер текущего поддерева (включая саму запись)

	for _, directoryEntry := range directoryEntries {
		if ctx.Err() != nil {
			return nil, 0, ctx.Err()
		}
		if directoryEntry.Type()&os.ModeSymlink != 0 {
			f.logger.LogAttrs(
				f.loggerCtx,
				slog.LevelWarn,
				"Обнаружена и пропущена символическая ссылка",
				slog.String("fullPath", filepath.Join(currentAbsolutePath, directoryEntry.Name())),
			)
			continue
		}

		entryName := directoryEntry.Name()
		nextAbsolutePath := filepath.Join(currentAbsolutePath, entryName)
		nextPath := filepath.Join(currentPath, entryName)

		entryInfo, err := directoryEntry.Info()
		if err != nil {
			return nil, 0, err
		}

		var resourceContentFunc func(string, string) (io.ReadCloser, error)
		if directoryEntry.IsDir() {
			resourceContentFunc = f.openDirContent
		} else {
			resourceContentFunc = f.openFileContent
		}

		resourceContent := cont.ResourceContent{
			FullPath:     nextAbsolutePath,
			RelativePath: nextPath,
			OpenContent:  resourceContentFunc,
		}

		hashValue, err := f.hashManager.ResolveHash(resourceContent, nextAbsolutePath)
		if err != nil {
			return nil, 0, err
		}

		fileMetadata := f.createMetadata(entryInfo)
		fileInfo, err := domain.NewFileInfo(fileMetadata, hashValue)
		if err != nil {
			return nil, 0, err
		}

		filePath, err := domain.NewPath(nextPath)
		if err != nil {
			return nil, 0, err
		}

		fileEntry, err := domain.NewFileEntry(filePath, subtreeSize, fileInfo)
		if err != nil {
			return nil, 0, err
		}

		if directoryEntry.IsDir() {
			// рекурсия — получаем список и размер поддерева детей
			nested, nestedSubtreeSize, err := f.collectFileEntriesRecursive(ctx, nextAbsolutePath, nextPath)
			if err != nil {
				return nil, 0, err
			}

			// теперь знаем полный размер поддерева этой директории
			fileEntry.SubtreeSize = 1 + nestedSubtreeSize

			collectedEntries = append(collectedEntries, fileEntry)
			collectedEntries = append(collectedEntries, nested...)

			subtreeSize += 1 + nestedSubtreeSize
		} else {
			// для обычного файла размер поддерева всегда 1
			fileEntry.SubtreeSize = 1

			collectedEntries = append(collectedEntries, fileEntry)
			subtreeSize += 1
		}
	}

	return collectedEntries, subtreeSize, nil
}

func (f *FileManager) createFileEntryForDirectory(
	ctx context.Context,
	absolutePath string,
	relativePath string,
	subtreeSize uint64,
) (domain.FileEntry, error) {
	if ctx.Err() != nil {
		return domain.FileEntry{}, ctx.Err()
	}

	resourceContent := cont.ResourceContent{
		OpenContent: f.openDirContent,
	}

	hashValue, err := f.hashManager.ResolveHash(resourceContent, absolutePath)
	if err != nil {
		return domain.FileEntry{}, err
	}

	info, err := f.fileSystem.Stat(absolutePath)
	if err != nil {
		return domain.FileEntry{}, err
	}

	metadata := f.createMetadata(info)
	fileInfo, err := domain.NewFileInfo(metadata, hashValue)
	if err != nil {
		return domain.FileEntry{}, err
	}

	filePath, err := domain.NewPath(relativePath)
	if err != nil {
		return domain.FileEntry{}, err
	}

	return domain.NewFileEntry(filePath, subtreeSize, fileInfo)
}
