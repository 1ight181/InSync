package base

import "insync/internal/domain"

func ToDomainSnapshot(bs BaseSnapshot) domain.Snapshot {
	files := make([]domain.FileEntry, len(bs.Files))

	for i, f := range bs.Files {
		files[i] = toDomainFileEntry(f)
	}

	return domain.NewSnapshot(files)
}

func toDomainFileEntry(f FileEntry) domain.FileEntry {
	return domain.FileEntry{
		RelativePath: domain.Path(f.RelativePath),
		SubtreeSize:  f.SubtreeSize,
		FileInfo:     toDomainFileInfo(f.FileInfo),
	}
}

func toDomainFileInfo(fi FileInfo) domain.FileInfo {
	return domain.FileInfo{
		Hash: fi.Hash,
		Metadata: domain.FileMetadata{
			IsDirectory:  fi.FileMetadata.IsDirectory,
			SizeBytes:    fi.FileMetadata.SizeBytes,
			ModifiedUnix: fi.FileMetadata.ModifiedUnix,
		},
	}
}
