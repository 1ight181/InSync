package domain

import (
	"errors"
	"slices"
	"strings"
)

type BaseSnapshot struct {
	Snapshot
	IsInitial bool
}

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

var (
	ErrBaseSnapshotNotFound = errors.New("base snapshot not found")
)
