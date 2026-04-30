package hash

import (
	"errors"
	"insync/internal/domain"
	"insync/internal/repository/sqlite/hash/dirty"
	"time"

	"gorm.io/gorm"
)

type HashRepository struct {
	db                           *gorm.DB
	hashCacheEntryExpireUnixTime int64
}

type HashRepositoryOptions struct {
	Db                           *gorm.DB
	HashCacheEntryExpireUnixTime int64
}

var (
	ErrInvalidHashRepositoryOptions = errors.New("Все поля HashRepositoryOptions должны быть заполнены")
)

func NewHashRepository(opts HashRepositoryOptions) (*HashRepository, error) {
	if opts.Db == nil {
		return nil, ErrInvalidHashRepositoryOptions
	}
	return &HashRepository{
		db:                           opts.Db,
		hashCacheEntryExpireUnixTime: opts.HashCacheEntryExpireUnixTime,
	}, nil
}

func (r *HashRepository) GetHashCache() (map[domain.Path]string, error) {
	var entries []HashCacheEntry
	if err := r.db.Find(&entries).Error; err != nil {
		return nil, err
	}
	var expired []HashCacheEntry
	cache := make(map[domain.Path]string, len(entries))
	for _, entry := range entries {
		if entry.Expires < time.Now().Unix() {
			expired = append(expired, entry)
		}

		// Уже валидированные при добавлении
		cache[domain.Path(entry.FullPath)] = entry.Hash
	}

	if len(expired) > 0 {
		if err := r.db.Delete(&expired).Error; err != nil {
			return nil, err
		}
	}

	return cache, nil
}

func (r *HashRepository) SetHashCache(fullPath domain.Path, hash string) error {
	var hashCacheEntry HashCacheEntry
	if err := r.db.First(&hashCacheEntry, "full_path = ?", fullPath).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hashCacheEntry.FullPath = fullPath.String()
	hashCacheEntry.Hash = hash
	hashCacheEntry.Expires = r.hashCacheEntryExpireUnixTime

	if err := r.db.Save(&hashCacheEntry).Error; err != nil {
		return err
	}

	return nil
}

func (r *HashRepository) GetDirtyPaths() (map[domain.ScopedPath]struct{}, error) {
	var dirtyPathEntries []dirty.DirtyPath
	if err := r.db.Find(&dirtyPathEntries).Error; err != nil {
		return nil, err
	}

	dirtyPaths := make(map[domain.ScopedPath]struct{})
	for _, dirtyPathEntry := range dirtyPathEntries {
		dirtyPaths[domain.ScopedPath{Root: domain.RootName(dirtyPathEntry.RootName), Path: domain.Path(dirtyPathEntry.FullPath)}] = struct{}{}
	}

	return dirtyPaths, nil
}

func (r *HashRepository) SetDirtyPath(scopedPath domain.ScopedPath) error {
	return r.db.Save(&dirty.DirtyPath{RootName: scopedPath.Root.String(), FullPath: scopedPath.Path.String()}).Error
}

func (r *HashRepository) RemoveDirtyPath(scopedPath domain.ScopedPath) error {
	return r.db.Delete(&dirty.DirtyPath{RootName: scopedPath.Root.String(), FullPath: scopedPath.Path.String()}).Error
}
