package server

import (
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
)

func domainFileEntriesToPb(fileEntries []domain.FileEntry) []*insyncpb.FileEntry {
	pbFileEntries := make([]*insyncpb.FileEntry, 0, len(fileEntries))
	for _, fileEntry := range fileEntries {
		pbFileEntries = append(pbFileEntries, domainFileEntryToPb(fileEntry))
	}
	return pbFileEntries
}

func domainFileEntryToPb(fileEntry domain.FileEntry) *insyncpb.FileEntry {
	return &insyncpb.FileEntry{
		RelativePath: fileEntry.RelativePath.String(),
		SizeBytes:    fileEntry.FileInfo.Metadata.SizeBytes,
		ModifiedUnix: fileEntry.FileInfo.Metadata.ModifiedUnix,
		Hash:         fileEntry.FileInfo.Hash,
		IsDirectory:  fileEntry.FileInfo.Metadata.IsDirectory,
		SubtreeSize:  fileEntry.SubtreeSize,
	}
}
