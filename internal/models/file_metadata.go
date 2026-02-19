package models

type FileMetadata struct {
	FileUuid     string
	RelativePath string
	SizeBytes    uint32
	ModifiedUnix uint32
	Hash         string
	IsDirectory  bool
}
