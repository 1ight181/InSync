package domain

import "errors"

type FileEntry struct {
	RelativePath Path
	FileInfo     FileInfo
}

var (
	ErrInvalidFileInfo = errors.New("FileInfo должен содержать валидные данные, включая непустой Hash, Path, RootName")
)

func NewFileEntry(relativePath Path, fileInfo FileInfo) (FileEntry, error) {
	if relativePath == "" || fileInfo.Hash == "" {
		return FileEntry{}, ErrInvalidFileInfo
	}

	return FileEntry{
		RelativePath: relativePath,
		FileInfo:     fileInfo,
	}, nil
}
