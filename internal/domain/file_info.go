package domain

import "errors"

type FileInfo struct {
	Metadata FileMetadata
	Hash     string
}

var (
	ErrInvalidHash = errors.New("Hash не может быть пустой строкой")
)

func NewFileInfo(metadata FileMetadata, hash string) (FileInfo, error) {
	if hash == "" {
		return FileInfo{}, ErrInvalidHash
	}
	return FileInfo{
		Metadata: metadata,
		Hash:     hash,
	}, nil
}
