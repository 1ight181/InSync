package domain

type SyncPlan struct {
	LocalChanges  []LocalChange
	RemoteChanges []RemoteChange
	Conflicts     []Conflict
}

type LocalChange SyncChange
type RemoteChange SyncChange

type SyncChange struct {
	OldRelativePath Path
	NewRelativePath Path
	ChangeType      SyncChangeType
	ModifiedUnix    uint64
}

func (c SyncChange) ToLocalChange() LocalChange {
	return LocalChange(c)
}

func (c SyncChange) ToRemoteChange() RemoteChange {
	return RemoteChange(c)
}

type SyncChangeType int

const (
	Create SyncChangeType = iota
	Delete
	Modify
	Rename
	Move
)
