package domain

type FileMetadata struct {
	ModifiedUnix uint32
	SizeBytes    uint32
	Hash         string
	IsDirectory  bool
}

func NewFileMetadata(modifiedUnix, sizeBytes uint32, isDirectory bool, hash string) FileMetadata {
	return FileMetadata{
		ModifiedUnix: modifiedUnix,
		SizeBytes:    sizeBytes,
		IsDirectory:  isDirectory,
		Hash:         hash,
	}
}

func (f *FileMetadata) SetHash(hash string) *FileMetadata {
	f.Hash = hash
	return f
}
