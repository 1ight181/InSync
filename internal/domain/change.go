package domain

type SyncChange struct {
	RootName        RootName
	NewRelativePath RelativePath
	OldRelativePath RelativePath
	ChangeType      SyncChangeType
}

type SyncChangeType int

const (
	Create SyncChangeType = iota
	Delete
	Rename
)
