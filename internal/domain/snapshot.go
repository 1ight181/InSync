package domain

import (
	"slices"
	"strings"
)

type Snapshot struct {
	UnixTime uint64
	Files    []FileEntry
}

func NewSnapshot(unixTime uint64, files []FileEntry) Snapshot {
	slices.SortFunc(files, func(a, b FileEntry) int {
		return strings.Compare(a.RelativePath.String(), b.RelativePath.String())
	})
	return Snapshot{
		UnixTime: unixTime,
		Files:    files,
	}
}
