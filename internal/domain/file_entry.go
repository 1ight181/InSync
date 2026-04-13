package domain

import "errors"

type FileEntry struct {
	RelativePath RelativePath
	FileInfo     FileInfo
}

var (
	ErrInvalidFileInfo = errors.New("FileInfo должен содержать валидные данные, включая непустой Hash, RelativePath, RootName")
)

func NewFileEntry(relativePath RelativePath, fileInfo FileInfo) (FileEntry, error) {
	if relativePath == "" || fileInfo.Hash == "" {
		return FileEntry{}, ErrInvalidFileInfo
	}

	return FileEntry{
		RelativePath: relativePath,
		FileInfo:     fileInfo,
	}, nil
}
