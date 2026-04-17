package hash

import (
	"insync/internal/domain"
	"sync"
)

type IHashCache interface {
	LoadHashCache(hashSet map[string]string) error
	GetHashCache(fullPath string) (string, error)
	SetHashCache(fullPath string, hash string)
}

type HashCache struct {
	mu sync.RWMutex
	// FullPath -> Hash
	cache map[string]string
}

func NewHashCache() *HashCache {
	return &HashCache{
		cache: make(map[string]string),
	}
}

func (rc *HashCache) LoadHashCache(hashSet map[string]string) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	for fullPath, hash := range hashSet {
		rc.cache[fullPath] = hash
	}
	return nil
}

func (rc *HashCache) GetHashCache(fullPath domain.Path) (string, error) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	hash, exists := rc.cache[fullPath.String()]
	if !exists {
		return "", HashCacheNotFoundError{FullPath: fullPath.String()}
	}
	return hash, nil
}

func (rc *HashCache) SetHashCache(fullPath domain.Path, hash string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.cache[fullPath.String()] = hash
}
