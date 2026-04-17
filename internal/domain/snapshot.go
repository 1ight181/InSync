package domain

import (
	"slices"
	"strings"
)

type Snapshot struct {
	Files []FileEntry
}

func NewSnapshot(files []FileEntry) Snapshot {
	slices.SortFunc(files, func(a, b FileEntry) int {
		return strings.Compare(a.RelativePath.String(), b.RelativePath.String())
	})
	return Snapshot{
		Files: files,
	}
}

type SnapshotMetadata struct {
	UnixTime uint64
	RootName RootName
	DeviceId DeviceId
}

type SnapshotWithMetadata struct {
	Snapshot Snapshot
	Metadata SnapshotMetadata
}
