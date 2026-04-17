package hash

import (
	"context"
	"insync/internal/domain"
	cont "insync/internal/infrastructure/filemanager/content"
	"log/slog"
)

type HashManager struct {
	hashCache      IHashCache
	hashCalculator IHashCalculator
	pathTreeReader IPathTreeReader
	dirtyPaths     map[domain.Path]struct{}

	logger    *slog.Logger
	loggerCtx context.Context
}

type HashManagerOptions struct {
	HashCache      IHashCache
	HashCalculator IHashCalculator
	PathTreeReader IPathTreeReader

	Logger    *slog.Logger
	LoggerCtx context.Context
}

func NewHashManager(options HashManagerOptions) *HashManager {
	return &HashManager{
		hashCache:      options.HashCache,
		hashCalculator: options.HashCalculator,
		pathTreeReader: options.PathTreeReader,

		dirtyPaths: make(map[domain.Path]struct{}),
		logger:     options.Logger,
		loggerCtx:  options.LoggerCtx,
	}
}

func (h *HashManager) ResolveHash(resourceContent cont.ResourceContent, fullPath domain.Path) (string, error) {
	if _, isDirty := h.dirtyPaths[fullPath]; !isDirty {
		if hash, err := h.hashCache.GetHashCache(fullPath); err == nil {
			return hash, nil
		}
		h.logger.LogAttrs(
			h.loggerCtx,
			slog.LevelDebug,
			"Кэш для хэша не найден",
			slog.String("fullPath", fullPath.String()),
		)
	} else {
		h.logger.LogAttrs(
			h.loggerCtx,
			slog.LevelDebug,
			"Путь является dirty",
			slog.String("fullPath", fullPath.String()),
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
		slog.String("fullPath", fullPath.String()),
		slog.String("hash", hash),
	)

	if err := h.hashCache.SetHashCache(fullPath, hash); err != nil {
		return "", err
	}

	h.logger.LogAttrs(
		h.loggerCtx,
		slog.LevelDebug,
		"Хэш успешно сохранен в кэш",
		slog.String("fullPath", fullPath.String()),
		slog.String("hash", hash),
	)

	return hash, nil
}

func (h *HashManager) MarkDirty(fullPath domain.Path) error {
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
