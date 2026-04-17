package client

import (
	"insync/internal/domain"
	"insync/internal/transport/grpc/insyncpb"

	"go.uber.org/multierr"
)

func pbSnapshotToDomain(pbSnapshot *insyncpb.Snapshot) (domain.Snapshot, error) {
	if pbSnapshot == nil {
		return domain.Snapshot{}, nil
	}

	files, err := pbFileEntriesToDomain(pbSnapshot.Files)
	if err != nil {
		return domain.Snapshot{}, err
	}

	snapshot := domain.NewSnapshot(files)

	return snapshot, nil
}

func pbFileEntriesToDomain(pbFileEntries []*insyncpb.FileEntry) ([]domain.FileEntry, error) {
	domainFileEntries := make([]domain.FileEntry, 0, len(pbFileEntries))
	var resultErr error
	for _, pbFileEntry := range pbFileEntries {
		fileEntry, err := pbFileEntryToDomain(pbFileEntry)
		if err != nil {
			resultErr = multierr.Append(resultErr, err)
			continue
		}

		domainFileEntries = append(domainFileEntries, fileEntry)
	}
	return domainFileEntries, resultErr
}

func pbFileEntryToDomain(pbFileEntry *insyncpb.FileEntry) (domain.FileEntry, error) {
	relativePath, err := domain.NewPath(pbFileEntry.RelativePath)
	if err != nil {
		return domain.FileEntry{}, err
	}
	fileMetadate := domain.FileMetadata{
		ModifiedUnix: pbFileEntry.ModifiedUnix,
		SizeBytes:    pbFileEntry.SizeBytes,
		IsDirectory:  pbFileEntry.IsDirectory,
	}

	fileInfo, err := domain.NewFileInfo(fileMetadate, pbFileEntry.Hash)
	if err != nil {
		return domain.FileEntry{}, err
	}

	fileEntry, err := domain.NewFileEntry(relativePath, pbFileEntry.SubtreeSize, fileInfo)
	if err != nil {
		return domain.FileEntry{}, err
	}

	return fileEntry, nil
}
