package interfaces

import "insync/internal/domain"

type IFileInfoCache interface {
	GetFileInfoCache(fullPath string, fileMetadata domain.FileMetadata) (domain.FileInfo, error)
	SetFileInfoCache(fullPath string, fileInfo domain.FileInfo)
}
