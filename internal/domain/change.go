package domain

type SyncChange struct {
	RootName        RootName
	OldRelativePath RelativePath
	NewRelativePath RelativePath
	ChangeType      SyncChangeType
}

type SyncChangeType int

const (
	Create SyncChangeType = iota
	Delete
	Rename
	Move
)
