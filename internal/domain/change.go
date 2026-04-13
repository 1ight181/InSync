package domain

type SyncChange struct {
	RootName        RootName
	OldRelativePath Path
	NewRelativePath Path
	ChangeType      SyncChangeType
}

type SyncChangeType int

const (
	Create SyncChangeType = iota
	Delete
	Rename
	Move
)
