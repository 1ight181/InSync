package server

import (
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
)

func DomainFileMetadataToPb(fileInfo domain.FileInfo) *insyncpb.FileMetadata {
	return &insyncpb.FileMetadata{
		RootName:     fileInfo.RootName,
		RelativePath: fileInfo.RelativePath,
		IsDirectory:  fileInfo.Metadata.IsDirectory,
		SizeBytes:    fileInfo.Metadata.SizeBytes,
		ModifiedUnix: fileInfo.Metadata.ModifiedUnix,
		Hash:         fileInfo.Metadata.Hash,
	}
}
