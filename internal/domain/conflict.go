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
	// Локальный файл удален, удаленный файл изменен
	ConflictLocalDeletedRemoteModified ConflictType = iota
	// Удаленный файл удален, локальный файл изменен
	ConflictRemoteDeletedLocalModified

	// Локальный файл перемещен, удаленный файл перемещен
	ConflictLocalMovedRemoteMoved
	// Локальный файл переименован, удаленный файл переименован
	ConflictLocalRenamedRemoteRenamed

	// Обе стороны изменили файл одновременно
	ConflictBothModifiedAtSameTime

	// Обе стороны создали файл с одинаковым именем, но разным содержимым
	ConflictBothCreatedAtSamePathConflict
)
