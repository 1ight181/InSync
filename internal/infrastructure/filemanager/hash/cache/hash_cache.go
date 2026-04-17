package cache

import (
	"insync/internal/domain"
	"sync"
)

type HashCache struct {
	mu sync.RWMutex
	// FullPath -> Hash
	cache  map[domain.Path]string
	loader IHashCacheLoader
}

type HashCacheOptions struct {
	HashCacheLoader IHashCacheLoader
}

func NewHashCache(opts HashCacheOptions) *HashCache {
	if opts.HashCacheLoader == nil {
		panic("Все поля HashCacheOptions должны быть заполнены")
	}

	hashCache := &HashCache{
		loader: opts.HashCacheLoader,
		cache:  make(map[domain.Path]string),
	}

	if err := hashCache.loadHashCache(); err != nil {
		panic(err)
	}

	return hashCache
}

func (rc *HashCache) loadHashCache() error {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	hashSet, err := rc.loader.GetHashCache()
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

func (rc *HashCache) SetHashCache(fullPath domain.Path, hash string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.cache[fullPath] = hash
}
