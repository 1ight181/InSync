package hash

import (
	"insync/internal/domain"

	"gorm.io/gorm"
)

type HashRepository struct {
	db *gorm.DB
}

type HashRepositoryOptions struct {
	Db *gorm.DB
}

func NewHashRepository(opts HashRepositoryOptions) *HashRepository {
	if opts.Db == nil {
		panic("все поля HashRepositoryOptions должны быть заполнены")
	}
	return &HashRepository{db: opts.Db}
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
