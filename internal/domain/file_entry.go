package domain

import "errors"

type FileEntry struct {
	RelativePath Path
	SubtreeSize  uint64
	FileInfo     FileInfo
}

var (
	ErrInvalidFileInfo = errors.New("FileInfo должен содержать валидные данные, включая непустой Hash, Path, RootName")
)

func NewFileEntry(relativePath Path, subtreeSize uint64, fileInfo FileInfo) (FileEntry, error) {
	if relativePath == "" || fileInfo.Hash == "" {
		return FileEntry{}, ErrInvalidFileInfo
	}

	return FileEntry{
		RelativePath: relativePath,
		SubtreeSize:  subtreeSize,
		FileInfo:     fileInfo,
	}, nil
}
