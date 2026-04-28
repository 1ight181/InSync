package hash

import (
	"errors"
	"insync/internal/domain"
	"insync/internal/repository/sqlite/hash/dirty"

	"gorm.io/gorm"
)

type HashRepository struct {
	db *gorm.DB
}

type HashRepositoryOptions struct {
	Db *gorm.DB
}

var (
	ErrInvalidHashRepositoryOptions = errors.New("Все поля HashRepositoryOptions должны быть заполнены")
)

func NewHashRepository(opts HashRepositoryOptions) (*HashRepository, error) {
	if opts.Db == nil {
		return nil, ErrInvalidHashRepositoryOptions
	}
	return &HashRepository{db: opts.Db}, nil
}

func (r *HashRepository) GetHashCache() (map[domain.Path]string, error) {
	var entries []HashCacheEntry
	if err := r.db.Find(&entries).Error; err != nil {
		return nil, err
	}

	cache := make(map[domain.Path]string, len(entries))
	for _, entry := range entries {
		// Уже валидированные при добавлении
		cache[domain.Path(entry.FullPath)] = entry.Hash
	}

	return cache, nil
}

func (r *HashRepository) SetHashCache(fullPath domain.Path, hash string) error {
	entry := &HashCacheEntry{
		FullPath: fullPath.String(),
		Hash:     hash,
	}

	if err := r.db.Save(entry).Error; err != nil {
		return err
	}

	return nil
}

func (r *HashRepository) GetDirtyPaths(rootName domain.RootName) (map[domain.Path]struct{}, error) {
	var entries []dirty.DirtyPath
	if err := r.db.Find(&entries, "root_name = ?", rootName).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	dirtyPaths := make(map[domain.Path]struct{}, len(entries))
	for _, entry := range entries {
		// Уже валидированные при добавлении
		dirtyPaths[domain.Path(entry.FullPath)] = struct{}{}
	}

	return dirtyPaths, nil
}
