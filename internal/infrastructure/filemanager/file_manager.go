package filemanager

import (
	"bytes"
	"context"
	"errors"
	"insync/internal/domain"
	cont "insync/internal/infrastructure/filemanager/content"
	pathtree "insync/internal/infrastructure/filemanager/pathtree"
	"sort"

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

	logger    *slog.Logger
	loggerCtx context.Context
}

type FileManagerOptions struct {
	RootResolver   IRootResolver
	HashManager    IHashManager
	FileSystem     IFileSystem
	PathTreeWriter IPathTreeWriter

	Logger *slog.Logger
}

var (
	ErrInvalidOpts = errors.New("Все поля FileManagerOptions должны быть заполнены")
)

func NewFileManager(opts FileManagerOptions) (*FileManager, error) {
	if opts.RootResolver == nil ||
		opts.HashManager == nil ||
		opts.FileSystem == nil ||
		opts.PathTreeWriter == nil ||
		opts.Logger == nil {
		return nil, ErrInvalidOpts
	}
	return &FileManager{
		rootResolver:   opts.RootResolver,
		hashManager:    opts.HashManager,
		fileSystem:     opts.FileSystem,
		pathTreeWriter: opts.PathTreeWriter,

		logger:    opts.Logger,
		loggerCtx: context.Background(),
	}, nil
}

func (f *FileManager) GetSnapshot(ctx context.Context, rootName domain.RootName, baseSnapshot *domain.BaseSnapshot) (domain.Snapshot, error) {
	if ctx.Err() != nil {
		return domain.Snapshot{}, ctx.Err()
	}

	scopedPath, err := domain.NewScopedPath(rootName, ".")
	if err != nil {
		return domain.Snapshot{}, err
	}
	resolvedRootPath, err := f.rootResolver.ResolveRoot(scopedPath)
	if err != nil {
		return domain.Snapshot{}, err
	}

	baseMetadataByPath := make(map[domain.Path]domain.FileMetadata)
	if baseSnapshot != nil {
		for _, entry := range baseSnapshot.Files {
			baseMetadataByPath[entry.RelativePath] = entry.FileInfo.Metadata
		}
	}

	shouldRecalculateHash := f.shouldRecalculateHash(ctx, baseMetadataByPath)
	if ctx.Err() != nil {
		return domain.Snapshot{}, ctx.Err()
	}

	allEntries, err := f.collectAllFileEntries(ctx, resolvedRootPath, scopedPath.Path, rootName, shouldRecalculateHash)
	if err != nil {
		return domain.Snapshot{}, err
	}

	snapshot := domain.NewSnapshot(allEntries)

	return snapshot, nil
}

func (f *FileManager) CreateDir(ctx context.Context, scopedPath domain.ScopedPath) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	resolvedPath, err := f.rootResolver.ResolveRoot(scopedPath)
	if err != nil {
		return err
	}

	if err := f.fileSystem.Mkdir(resolvedPath, 0755); err != nil {
		return err
	}

	return f.pathTreeWriter.AddPath(scopedPath)
}

func (f *FileManager) RenameFile(ctx context.Context, oldScopedPath domain.ScopedPath, newScopedPath domain.ScopedPath) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	oldResolvedPath, err := f.rootResolver.ResolveRoot(oldScopedPath)
	if err != nil {
		return err
	}

	newResolvedPath, err := f.rootResolver.ResolveRoot(newScopedPath)
	if err != nil {
		return err
	}

	if err := f.fileSystem.Rename(oldResolvedPath, newResolvedPath); err != nil {
		return err
	}

	if err := f.pathTreeWriter.RemovePath(oldScopedPath); err != nil && !errors.Is(err, pathtree.ErrNotFound) {
		return err
	}

	if err := f.pathTreeWriter.AddPath(newScopedPath); err != nil {
		return err
	}

	return f.hashManager.MarkDirty(newScopedPath)
}

func (f *FileManager) DeleteFile(ctx context.Context, scopedPath domain.ScopedPath) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	resolvedPath, err := f.rootResolver.ResolveRoot(scopedPath)
	if err != nil {
		return err
	}

	if err := f.fileSystem.Remove(resolvedPath); err != nil {
		return err
	}

	if err := f.pathTreeWriter.RemovePath(scopedPath); err != nil && !errors.Is(err, pathtree.ErrNotFound) {
		return err
	}

	return f.hashManager.MarkDirty(scopedPath)
}

func (f *FileManager) GetFile(ctx context.Context, scopedPath domain.ScopedPath) (io.ReadCloser, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	resolvedFullPath, err := f.rootResolver.ResolveRoot(scopedPath)
	if err != nil {
		return nil, err
	}

	return f.openFileContent(resolvedFullPath, scopedPath.Path)
}

func (f *FileManager) PutFile(ctx context.Context, scopedPath domain.ScopedPath, content io.Reader) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	resolvedPath, err := f.rootResolver.ResolveRoot(scopedPath)
	if err != nil {
		return err
	}

	destDir := resolvedPath.Dir()
	if err := os.MkdirAll(destDir.String(), 0755); err != nil {
		return err
	}

	if err := f.fileSystem.AtomicWrite(resolvedPath, content); err != nil {
		return err
	}

	if err := f.pathTreeWriter.AddPath(scopedPath); err != nil {
		return err
	}

	return f.hashManager.MarkDirty(scopedPath)
}

func (f *FileManager) openDirContent(fullPath, relativePath domain.Path) (io.ReadCloser, error) {

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
				slog.String("fullPath", fullPath.String()),
				slog.String("symlinkName", filepath.Join(fullPath.String(), entry.Name())),
			)
			continue
		}

		if entry.IsDir() {
			subDirFullPath, err := fullPath.Join(entry.Name())
			if err != nil {
				return nil, err
			}

			subDirPath, err := relativePath.Join(entry.Name())
			if err != nil {
				return nil, err
			}

			subDirContent, err := f.openDirContent(subDirFullPath, subDirPath)
			if err != nil {
				f.logger.LogAttrs(
					f.loggerCtx,
					slog.LevelError,
					"Ошибка при чтении директории",
					slog.String("fullPath", subDirFullPath.String()),
				)

				NewMultiCloser(closers...).Close()

				return nil, err
			}

			readers = append(readers, subDirContent)
			closers = append(closers, subDirContent)
			continue
		}

		fileFullPath, err := fullPath.Join(entry.Name())
		if err != nil {
			return nil, err
		}

		filePath, err := relativePath.Join(entry.Name())
		if err != nil {
			return nil, err
		}

		fileContent, err := f.openFileContentWithHeader(fileFullPath, filePath)
		if err != nil {
			f.logger.LogAttrs(
				f.loggerCtx,
				slog.LevelError,
				"Ошибка при чтении файла",
				slog.String("fullPath", fileFullPath.String()),
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

func (f *FileManager) openFileContentWithHeader(fullPath, relativePath domain.Path) (io.ReadCloser, error) {
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

func (f *FileManager) openFileContent(fullPath, _ domain.Path) (io.ReadCloser, error) {
	return f.fileSystem.Open(fullPath)
}

func (f *FileManager) createMetadata(entryInfo fs.FileInfo) domain.FileMetadata {
	return domain.FileMetadata{
		ModifiedUnix: uint64(entryInfo.ModTime().Unix()),
		SizeBytes:    uint64(entryInfo.Size()),
		IsDirectory:  entryInfo.IsDir(),
	}
}

func (f *FileManager) resolveHash(
	resourceContent cont.ResourceContent,
	rootName domain.RootName,
	shouldRecalculateHash bool,
) (string, error) {
	if shouldRecalculateHash {
		return f.hashManager.ResolveWithForceRecalc(resourceContent, rootName)
	}

	return f.hashManager.ResolveHash(resourceContent, rootName)
}

func (f *FileManager) collectAllFileEntries(
	ctx context.Context,
	rootAbsolutePath domain.Path,
	rootRelativePath domain.Path,
	rootName domain.RootName,
	shouldRecalculateHash bool,
) ([]domain.FileEntry, error) {

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	children, childrenSubtreeSize, err := f.collectFileEntriesRecursive(ctx, rootAbsolutePath, rootRelativePath, rootName, shouldRecalculateHash)
	if err != nil {
		return nil, err
	}

	rootEntrySubtreeSize := uint64(1 + childrenSubtreeSize)

	rootEntry, err := f.createFileEntryForDirectory(ctx, rootAbsolutePath, rootRelativePath, rootEntrySubtreeSize, rootName, shouldRecalculateHash)
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
	currentAbsolutePath domain.Path,
	currentPath domain.Path,
	rootName domain.RootName,
	shouldRecalculateHash bool,
) ([]domain.FileEntry, uint64, error) {
	directoryEntries, err := f.fileSystem.ReadDir(currentAbsolutePath)
	if err != nil {
		return nil, 0, err
	}

	var collectedEntries []domain.FileEntry
	var subtreeSize uint64 = 0

	for _, directoryEntry := range directoryEntries {
		if ctx.Err() != nil {
			return nil, 0, ctx.Err()
		}

		fullPath, err := currentAbsolutePath.Join(directoryEntry.Name())
		if err != nil {
			return nil, 0, err
		}

		if directoryEntry.Type()&os.ModeSymlink != 0 {
			f.logger.LogAttrs(
				f.loggerCtx,
				slog.LevelWarn,
				"Обнаружена и пропущена символическая ссылка",
				slog.String("fullPath", fullPath.String()),
			)
			continue
		}

		entryName := directoryEntry.Name()

		nextAbsolutePath, err := currentAbsolutePath.Join(entryName)
		if err != nil {
			return nil, 0, err
		}

		nextPath, err := currentPath.Join(entryName)
		if err != nil {
			return nil, 0, err
		}

		entryInfo, err := directoryEntry.Info()
		if err != nil {
			return nil, 0, err
		}

		var resourceContentFunc func(domain.Path, domain.Path) (io.ReadCloser, error)
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

		fileMetadata := f.createMetadata(entryInfo)
		hashValue, err := f.resolveHash(resourceContent, rootName, shouldRecalculateHash)
		if err != nil {
			return nil, 0, err
		}
		fileInfo, err := domain.NewFileInfo(fileMetadata, hashValue)
		if err != nil {
			return nil, 0, err
		}

		fileEntry, err := domain.NewFileEntry(nextPath, subtreeSize, fileInfo)
		if err != nil {
			return nil, 0, err
		}

		if directoryEntry.IsDir() {
			nested, nestedSubtreeSize, err := f.collectFileEntriesRecursive(ctx, nextAbsolutePath, nextPath, rootName, shouldRecalculateHash)
			if err != nil {
				return nil, 0, err
			}

			fileEntry.SubtreeSize = 1 + nestedSubtreeSize

			collectedEntries = append(collectedEntries, fileEntry)
			collectedEntries = append(collectedEntries, nested...)

			subtreeSize += 1 + nestedSubtreeSize
		} else {
			fileEntry.SubtreeSize = 1

			collectedEntries = append(collectedEntries, fileEntry)
			subtreeSize += 1
		}
	}

	return collectedEntries, subtreeSize, nil
}

func (f *FileManager) createFileEntryForDirectory(
	ctx context.Context,
	absolutePath domain.Path,
	relativePath domain.Path,
	subtreeSize uint64,
	rootName domain.RootName,
	shouldRecalculateHash bool,
) (domain.FileEntry, error) {
	if ctx.Err() != nil {
		return domain.FileEntry{}, ctx.Err()
	}

	resourceContent := cont.ResourceContent{
		FullPath:     absolutePath,
		RelativePath: relativePath,
		OpenContent:  f.openDirContent,
	}

	info, err := f.fileSystem.Stat(absolutePath)
	if err != nil {
		return domain.FileEntry{}, err
	}

	metadata := f.createMetadata(info)
	hashValue, err := f.resolveHash(resourceContent, rootName, shouldRecalculateHash)
	if err != nil {
		return domain.FileEntry{}, err
	}
	fileInfo, err := domain.NewFileInfo(metadata, hashValue)
	if err != nil {
		return domain.FileEntry{}, err
	}

	return domain.NewFileEntry(relativePath, subtreeSize, fileInfo)
}

func (f *FileManager) shouldRecalculateHash(ctx context.Context, baseMetadataByPath map[domain.Path]domain.FileMetadata) bool {
	if ctx.Err() != nil {
		return false
	}

	for path, metadata := range baseMetadataByPath {
		actualInfo, err := f.fileSystem.Stat(path)
		if err != nil {
			return true
		}

		actualMetadata := f.createMetadata(actualInfo)
		if actualMetadata != metadata {
			return true
		}

		if ctx.Err() != nil {
			return false
		}
	}

	return false
}
