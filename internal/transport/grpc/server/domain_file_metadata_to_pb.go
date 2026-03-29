package server

import (
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
)

func DomainFileEntryToPb(fileEntry domain.FileEntry) *insyncpb.FileEntry {
	return &insyncpb.FileEntry{
		RootName:     fileEntry.RootName,
		RelativePath: fileEntry.RelativePath,
		IsDirectory:  fileEntry.FileInfo.Metadata.IsDirectory,
		SizeBytes:    fileEntry.FileInfo.Metadata.SizeBytes,
		ModifiedUnix: fileEntry.FileInfo.Metadata.ModifiedUnix,
		Hash:         fileEntry.FileInfo.Hash,
	}
}
