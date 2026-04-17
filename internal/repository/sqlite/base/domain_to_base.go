package base

import (
	"insync/internal/domain"
)

func ToBaseSnapshot(s domain.SnapshotWithMetadata) BaseSnapshot {
	baseSnap := BaseSnapshot{
		UnixTime:       s.Metadata.UnixTime,
		RootName:       s.Metadata.RootName.String(),
		LocalDeviceId:  s.Metadata.DeviceId.String(),
		RemoteDeviceId: s.Metadata.DeviceId.String(),
		Files:          make([]FileEntry, len(s.Snapshot.Files)),
	}

	for i, file := range s.Snapshot.Files {
		baseSnap.Files[i] = toBaseFileEntry(file)
	}

	return baseSnap
}

func toBaseFileEntry(f domain.FileEntry) FileEntry {
	return FileEntry{
		RelativePath: f.RelativePath.String(),
		SubtreeSize:  f.SubtreeSize,

		FileInfo: toBaseFileInfo(f.FileInfo),
	}
}

func toBaseFileInfo(fi domain.FileInfo) FileInfo {
	return FileInfo{
		Hash: fi.Hash,
		FileMetadata: FileMetadata{
			IsDirectory:  fi.Metadata.IsDirectory,
			SizeBytes:    fi.Metadata.SizeBytes,
			ModifiedUnix: fi.Metadata.ModifiedUnix,
		},
	}
}
