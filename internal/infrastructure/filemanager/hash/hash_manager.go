package hash

import (
	"context"
	"errors"
	"insync/internal/domain"
	cont "insync/internal/infrastructure/filemanager/content"
	"log/slog"
)

type HashManager struct {
	hashCache      IHashCache
	hashCalculator IHashCalculator
	pathTreeReader IPathTreeReader
	dirtyPaths     map[domain.ScopedPath]struct{}

	logger    *slog.Logger
	loggerCtx context.Context
}

type HashManagerOptions struct {
	HashCache            IHashCache
	HashCalculator       IHashCalculator
	PathTreeReader       IPathTreeReader
	DirtyPathsRepository IDirtyPathsRepository

	Logger *slog.Logger
}

var (
	ErrInvalidOpts = errors.New("Все поля HashManagerOptions должны быть заполнены")
)

func NewHashManager(opts HashManagerOptions) (*HashManager, error) {
	if opts.HashCache == nil ||
		opts.HashCalculator == nil ||
		opts.PathTreeReader == nil ||
		opts.Logger == nil {
		return nil, ErrInvalidOpts
	}

	dirtyPaths, err := opts.DirtyPathsRepository.GetDirtyPaths()
	if err != nil {
		return nil, err
	}

	return &HashManager{
		hashCache:      opts.HashCache,
		hashCalculator: opts.HashCalculator,
		pathTreeReader: opts.PathTreeReader,

		dirtyPaths: dirtyPaths,
		logger:     opts.Logger,
		loggerCtx:  context.Background(),
	}, nil
}

func (h *HashManager) ResolveHash(resourceContent cont.ResourceContent, rootName domain.RootName) (string, error) {
	// TODO: сделать отдельно rel и abs пути, сделать через конструктор
	scopedPath := domain.ScopedPath{
		Root: rootName,
		Path: resourceContent.FullPath,
	}
	if _, isDirty := h.dirtyPaths[scopedPath]; !isDirty {
		if hash, err := h.hashCache.GetHashCache(resourceContent.FullPath); err == nil {
			return hash, nil
		}
		h.logger.LogAttrs(
			h.loggerCtx,
			slog.LevelDebug,
			"Кэш для хэша не найден",
			slog.String("fullPath", resourceContent.FullPath.String()),
		)
	} else {
		h.logger.LogAttrs(
			h.loggerCtx,
			slog.LevelDebug,
			"Путь является dirty",
			slog.String("fullPath", resourceContent.FullPath.String()),
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
		slog.String("fullPath", resourceContent.FullPath.String()),
		slog.String("hash", hash),
	)

	if err := h.hashCache.SetHashCache(resourceContent.FullPath, hash); err != nil {
		return "", err
	}

	h.logger.LogAttrs(
		h.loggerCtx,
		slog.LevelDebug,
		"Хэш успешно сохранен в кэш",
		slog.String("fullPath", resourceContent.FullPath.String()),
		slog.String("hash", hash),
	)

	return hash, nil
}

func (h *HashManager) MarkDirty(scopedPath domain.ScopedPath) error {
	h.dirtyPaths[scopedPath] = struct{}{}

	parents, err := h.pathTreeReader.GetParents(scopedPath)
	if err != nil {
		return err
	}

	for _, parent := range parents {
		h.dirtyPaths[parent] = struct{}{}
	}

	return nil
}
