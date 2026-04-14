package domain

type FileMetadata struct {
	ModifiedUnix uint64
	SizeBytes    uint64
	IsDirectory  bool
}
