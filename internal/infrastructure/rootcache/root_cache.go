package rootcache

import "sync"

type RootCache struct {
	mu    sync.RWMutex
	cache map[string]string
}

func NewRootCache() *RootCache {
	return &RootCache{
		cache: make(map[string]string),
	}
}

func (rc *RootCache) GetRootCache(rootName string) (string, error) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	path, exists := rc.cache[rootName]
	if !exists {
		return "", RootCacheNotFoundError{RootName: rootName}
	}
	return path, nil
}

func (rc *RootCache) SetRootCache(rootName string, path string) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.cache[rootName] = path
	return nil
}
