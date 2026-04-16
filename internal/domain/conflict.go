package domain

type Conflict struct {
	LocalRelativePath  Path
	RemoteRelativePath Path
	BaseRelativePath   Path

	LocalModifiedUnix  uint64
	RemoteModifiedUnix uint64
	BaseModifiedUnix   uint64

	Conflict ConflictType
}

type ConflictType int

const (
	ConflictLocalDeletedRemoteModified ConflictType = iota
	ConflictRemoteDeletedLocalModified

	ConflictLocalMovedRemoteMoved
	ConflictLocalRenamedRemoteRenamed

	ConflictBothModifiedConflictAtSameTime
	ConflictBothCreatedAtSamePathConflict
)
