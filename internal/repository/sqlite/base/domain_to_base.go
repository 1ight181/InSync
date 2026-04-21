package base

import (
	"insync/internal/domain"
)

func ToBaseSnapshot(snapshot domain.Snapshot, localDeviceId, remoteDeviceId domain.DeviceId, rootName domain.RootName) BaseSnapshot {
	baseSnap := BaseSnapshot{
		RootName:       rootName.String(),
		LocalDeviceId:  localDeviceId.String(),
		RemoteDeviceId: remoteDeviceId.String(),
		Files:          make([]FileEntry, len(snapshot.Files)),
	}

	for i, file := range snapshot.Files {
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
