package hash

import (
	"context"
	cont "insync/internal/infrastructure/filemanager/content"
	"log/slog"
)

type HashManager struct {
	hashCache      IHashCache
	hashCalculator IHashCalculator
	pathTreeReader IPathTreeReader
	dirtyPaths     map[string]struct{}

	logger    *slog.Logger
	loggerCtx context.Context
}

type HashManagerOptions struct {
	hashCache      IHashCache
	HashCalculator IHashCalculator
	PathTreeReader IPathTreeReader

	Logger    *slog.Logger
	LoggerCtx context.Context
}

func NewHashManager(options HashManagerOptions) *HashManager {
	return &HashManager{
		hashCache:      options.hashCache,
		hashCalculator: options.HashCalculator,
		pathTreeReader: options.PathTreeReader,

		dirtyPaths: make(map[string]struct{}),
		logger:     options.Logger,
		loggerCtx:  options.LoggerCtx,
	}
}

func (h *HashManager) ResolveHash(resourceContent cont.ResourceContent) (string, error) {
	fullPath := resourceContent.Path

	if _, isDirty := h.dirtyPaths[fullPath]; !isDirty {
		if hash, err := h.hashCache.GetHashCache(fullPath); err == nil {
			return hash, nil
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

	h.hashCache.SetHashCache(fullPath, hash)

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
