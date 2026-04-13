package hash

import (
	"context"
	"insync/internal/domain"
	cont "insync/internal/infrastructure/filemanager/content"
	"log/slog"
)

type HashManager struct {
	fileInfoCache  IFileInfoCache
	hashCalculator IHashCalculator
	pathTreeReader IPathTreeReader
	dirtyPaths     map[string]struct{}

	logger    *slog.Logger
	loggerCtx context.Context
}

type HashManagerOptions struct {
	FileInfoCache  IFileInfoCache
	HashCalculator IHashCalculator
	PathTreeReader IPathTreeReader

	Logger    *slog.Logger
	LoggerCtx context.Context
}

func NewHashManager(options HashManagerOptions) *HashManager {
	return &HashManager{
		fileInfoCache:  options.FileInfoCache,
		hashCalculator: options.HashCalculator,
		pathTreeReader: options.PathTreeReader,

		dirtyPaths: make(map[string]struct{}),
		logger:     options.Logger,
		loggerCtx:  options.LoggerCtx,
	}
}

func (h *HashManager) ResolveHash(resourceContent cont.ResourceContent, fileMetadata domain.FileMetadata) (string, error) {
	fullPath := resourceContent.Path

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
