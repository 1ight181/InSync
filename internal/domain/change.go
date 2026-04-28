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

func (c SyncChange) SortPath() Path {
	if c.ChangeType == Delete {
		return c.OldRelativePath
	}
	return c.NewRelativePath
}

type SyncChangeType int

const (
	CreateFile SyncChangeType = iota
	CreateDir
	Delete
	Modify
	Rename
	Move
)
