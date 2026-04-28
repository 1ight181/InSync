package server

import (
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"
)

func pbSnapshotToDomain(snapshot *insyncpb.Snapshot) domain.Snapshot {
	return domain.Snapshot{
		Files: pbFileEntriesToDomain(snapshot.Files),
	}
}

func pbFileEntriesToDomain(pbFileEntries []*insyncpb.FileEntry) []domain.FileEntry {
	fileEntries := make([]domain.FileEntry, len(pbFileEntries))
	for i, pbFileEntry := range pbFileEntries {
		fileEntries[i] = pbFileEntryToDomain(pbFileEntry)
	}
	return fileEntries
}

func pbFileEntryToDomain(pbFileEntry *insyncpb.FileEntry) domain.FileEntry {
	return domain.FileEntry{
		// Валидность гарантируется контрактом
		RelativePath: domain.Path(pbFileEntry.RelativePath),
		FileInfo: domain.FileInfo{
			Hash: pbFileEntry.Hash,
			Metadata: domain.FileMetadata{
				SizeBytes:    pbFileEntry.SizeBytes,
				ModifiedUnix: pbFileEntry.ModifiedUnix,
				IsDirectory:  pbFileEntry.IsDirectory,
			},
		},
		SubtreeSize: pbFileEntry.SubtreeSize,
	}
}
