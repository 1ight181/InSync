package domain

type SyncChange struct {
	RootName        RootName
	OldRelativePath Path
	NewRelativePath Path
	ChangeType      SyncChangeType
	ModifiedUnix    uint64
}

type SyncChangeType int

const (
	Created SyncChangeType = iota
	Deleted
	Modified
	Renamed
	Moved
)
