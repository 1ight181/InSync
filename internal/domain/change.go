package domain

type LocalChange SyncChange

func (c LocalChange) ToSyncChange() SyncChange {
	return SyncChange(c)
}

type RemoteChange SyncChange

func (c RemoteChange) ToSyncChange() SyncChange {
	return SyncChange(c)
}

type SyncChange struct {
	OldRelativePath Path
	NewRelativePath Path
	ChangeType      SyncChangeType
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
