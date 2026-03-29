package domain

import "errors"

type FileEntry struct {
	RootName     string
	RelativePath string
	FileInfo     FileInfo
}

var (
	ErrInvalidFileInfo = errors.New("FileInfo должен содержать валидные данные, включая непустой Hash, RelativePath, RootName")
)

func NewFileEntry(rootName, relativePath string, fileInfo FileInfo) (FileEntry, error) {
	if relativePath == "" || rootName == "" || fileInfo.Hash == "" {
		return FileEntry{}, ErrInvalidFileInfo
	}

	return FileEntry{
		RootName:     rootName,
		RelativePath: relativePath,
		FileInfo:     fileInfo,
	}, nil
}
