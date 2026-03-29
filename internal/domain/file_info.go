package domain

import "errors"

type FileInfo struct {
	RootName     string
	RelativePath string
	Metadata     FileMetadata
}

var (
	ErrInvalidFileInfo = errors.New("FileInfo должен содержать валидные данные, включая непустой Hash, RelativePath, RootName")
)

func NewFileInfo(rootName, relativePath string, metadata FileMetadata) (FileInfo, error) {
	if metadata.Hash == "" ||
		relativePath == "" ||
		rootName == "" {
		return FileInfo{}, ErrInvalidFileInfo
	}

	return FileInfo{
		RootName:     rootName,
		RelativePath: relativePath,
		Metadata:     metadata,
	}, nil
}
