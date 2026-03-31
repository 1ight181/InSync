package hash

import (
	"insync/internal/domain"
	"insync/internal/interfaces"
	"sync"
)

type FileInfoCache struct {
	mu sync.RWMutex
	// FullPath -> Hash
	cache map[string]domain.FileInfo
}

func NewHashCache() interfaces.IFileInfoCache {
	return &FileInfoCache{
		cache: make(map[string]domain.FileInfo),
	}
}

func (rc *FileInfoCache) GetFileInfoCache(fullPath string, fileMetadata domain.FileMetadata) (domain.FileInfo, error) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	fileInfo, exists := rc.cache[fullPath]
	if !exists {
		return domain.FileInfo{}, FileInfoCacheNotFoundError{FullPath: fullPath}
	}
	return fileInfo, nil
}

func (rc *FileInfoCache) SetFileInfoCache(fullPath string, fileInfo domain.FileInfo) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.cache[fullPath] = fileInfo
}
