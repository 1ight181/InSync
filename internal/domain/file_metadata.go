package domain

type FileMetadata struct {
	ModifiedUnix uint32
	SizeBytes    uint32
	IsDirectory  bool
}
