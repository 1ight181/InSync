package scan

import "insync/internal/domain"

type IFileManager interface {
	GetFileList(rootName domain.RootName) ([]domain.FileEntry, error)
}
