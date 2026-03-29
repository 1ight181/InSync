package domain

type FileMetadata struct {
	RootName     string
	RelativePath string
	SizeBytes    uint32
	ModifiedUnix uint32
	Hash         string
	IsDirectory  bool
}
