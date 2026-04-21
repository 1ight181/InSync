package cache

import (
	"errors"
	"insync/internal/domain"
	"sync"
)

type HashCache struct {
	mu sync.RWMutex
	// FullPath -> Hash
	cache               map[domain.Path]string
	hashCacheRepository IHashCacheRepository
}

type HashCacheOptions struct {
	HashCacheRepository IHashCacheRepository
}

var (
	ErrInvalidOpts = errors.New("Все поля HashCacheOptions должны быть заполнены")
)

func NewHashCache(opts HashCacheOptions) (*HashCache, error) {
	if opts.HashCacheRepository == nil {
		return nil, ErrInvalidOpts
	}

	hashCache := &HashCache{
		hashCacheRepository: opts.HashCacheRepository,
		cache:               make(map[domain.Path]string),
	}

	if err := hashCache.loadHashCache(); err != nil {
		return nil, err
	}

	return hashCache, nil
}

func (rc *HashCache) loadHashCache() error {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	hashSet, err := rc.hashCacheRepository.GetHashCache()
	if err != nil {
		return err
	}
	for fullPath, hash := range hashSet {
		rc.cache[fullPath] = hash
	}

	return nil
}

func (rc *HashCache) GetHashCache(fullPath domain.Path) (string, error) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	hash, exists := rc.cache[fullPath]
	if !exists {
		return "", HashCacheNotFoundError{FullPath: fullPath.String()}
	}
	return hash, nil
}

func (rc *HashCache) SetHashCache(fullPath domain.Path, hash string) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	if err := rc.hashCacheRepository.SetHashCache(fullPath, hash); err != nil {
		return err
	}

	rc.cache[fullPath] = hash

	return nil
}
