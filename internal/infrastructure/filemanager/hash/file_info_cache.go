package hash

import (
	"insync/internal/domain"
	"sync"
)

type IFileInfoCache interface {
	LoadFileInfo(fileInfoSet map[string]domain.FileInfo) error
	GetFileInfoCache(fullPath string, fileMetadata domain.FileMetadata) (domain.FileInfo, error)
	SetFileInfoCache(fullPath string, fileInfo domain.FileInfo)
}

type FileInfoCache struct {
	mu sync.RWMutex
	// FullPath -> Hash
	cache map[string]domain.FileInfo
}

func NewFileInfoCache() *FileInfoCache {
	return &FileInfoCache{
		cache: make(map[string]domain.FileInfo),
	}
}

func (rc *FileInfoCache) LoadFileInfo(fileInfoSet map[string]domain.FileInfo) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	for fullPath, fileInfo := range fileInfoSet {
		rc.cache[fullPath] = fileInfo
	}
	return nil
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
