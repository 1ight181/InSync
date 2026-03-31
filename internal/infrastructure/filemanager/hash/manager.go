package hash

import (
	"context"
	"insync/internal/domain"
	"insync/internal/interfaces"
	"log/slog"
)

type HashManager struct {
	fileInfoCache  interfaces.IFileInfoCache
	hashCalculator interfaces.IHashCalculator
	pathTreeReader interfaces.IPathTreeReader
	dirtyPaths     map[string]struct{}

	logger    *slog.Logger
	loggerCtx context.Context
}

type HashManagerOptions struct {
	FileInfoCache  interfaces.IFileInfoCache
	HashCalculator interfaces.IHashCalculator
	PathTreeReader interfaces.IPathTreeReader

	Logger    *slog.Logger
	LoggerCtx context.Context
}

func NewHashManager(options HashManagerOptions) interfaces.IHashManager {
	return &HashManager{
		fileInfoCache:  options.FileInfoCache,
		hashCalculator: options.HashCalculator,
		pathTreeReader: options.PathTreeReader,

		dirtyPaths: make(map[string]struct{}),
		logger:     options.Logger,
		loggerCtx:  options.LoggerCtx,
	}
}

func (h *HashManager) ResolveHash(fullPath string, fileMetadata domain.FileMetadata, getContent func(string) ([]byte, error)) (string, error) {
	if _, isDirty := h.dirtyPaths[fullPath]; !isDirty {
		if fileInfo, err := h.fileInfoCache.GetFileInfoCache(fullPath, fileMetadata); err == nil {
			return fileInfo.Hash, nil
		}
		h.logger.LogAttrs(
			h.loggerCtx,
			slog.LevelDebug,
			"Кэш для хэша не найден",
			slog.String("fullPath", fullPath),
		)
	} else {
		h.logger.LogAttrs(
			h.loggerCtx,
			slog.LevelDebug,
			"Путь является dirty",
			slog.String("fullPath", fullPath),
		)
	}

	content, err := getContent(fullPath)
	if err != nil {
		return "", err
	}

	resourceContent := domain.ResourceContent{
		Content: content,
		Path:    fullPath,
	}

	hash, err := h.hashCalculator.CalculateHash(resourceContent)
	if err != nil {
		return "", err
	}

	h.logger.LogAttrs(
		h.loggerCtx,
		slog.LevelDebug,
		"Хэш успешно рассчитан",
		slog.String("fullPath", fullPath),
		slog.String("hash", hash),
	)

	fileInfo := domain.FileInfo{
		Hash:     hash,
		Metadata: fileMetadata,
	}

	h.fileInfoCache.SetFileInfoCache(fullPath, fileInfo)

	h.logger.LogAttrs(
		h.loggerCtx,
		slog.LevelDebug,
		"Хэш успешно сохранен в кэш",
		slog.String("fullPath", fullPath),
		slog.String("hash", hash),
	)

	return hash, nil
}

func (h *HashManager) MarkDirty(fullPath string) error {
	h.dirtyPaths[fullPath] = struct{}{}

	parents, err := h.pathTreeReader.GetParents(fullPath)
	if err != nil {
		return err
	}

	for _, parent := range parents {
		h.dirtyPaths[parent] = struct{}{}
	}

	return nil
}
