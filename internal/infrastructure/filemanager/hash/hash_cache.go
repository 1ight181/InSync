package hash

import (
	"sync"
)

type IHashCache interface {
	LoadHashCache(fileInfoSet map[string]string) error
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

func (rc *HashCache) LoadHashCache(fileInfoSet map[string]string) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	for fullPath, fileInfo := range fileInfoSet {
		rc.cache[fullPath] = fileInfo
	}
	return nil
}

func (rc *HashCache) GetHashCache(fullPath string) (string, error) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	fileInfo, exists := rc.cache[fullPath]
	if !exists {
		return "", HashCacheNotFoundError{FullPath: fullPath}
	}
	return fileInfo, nil
}

func (rc *HashCache) SetHashCache(fullPath string, hash string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.cache[fullPath] = hash
}
