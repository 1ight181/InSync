package server

import (
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
)

func DomainFileMetadataToPb(fileMetadata domain.FileMetadata) *insyncpb.FileMetadata {
	return &insyncpb.FileMetadata{
		RootName:     fileMetadata.RootName,
		RelativePath: fileMetadata.RelativePath,
		IsDirectory:  fileMetadata.IsDirectory,
		SizeBytes:    fileMetadata.SizeBytes,
		ModifiedUnix: fileMetadata.ModifiedUnix,
		Hash:         fileMetadata.Hash,
	}
}
