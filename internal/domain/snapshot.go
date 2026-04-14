package domain

import (
	"slices"
	"strings"
)

type Snapshot struct {
	RootName RootName
	UnixTime uint64
	Files    []FileEntry
}

func NewSnapshot(rootName RootName, unixTime uint64, files []FileEntry) Snapshot {
	slices.SortFunc(files, func(a, b FileEntry) int {
		return strings.Compare(a.RelativePath.String(), b.RelativePath.String())
	})
	return Snapshot{
		RootName: rootName,
		UnixTime: unixTime,
		Files:    files,
	}
}
