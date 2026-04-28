package domain

import "sort"

type SyncPlan struct {
	LocalChanges  []LocalChange
	RemoteChanges []RemoteChange
	Conflicts     []Conflict
}

func (s SyncPlan) IsEmpty() bool {
	return len(s.LocalChanges) == 0 && len(s.RemoteChanges) == 0 && len(s.Conflicts) == 0
}
func (s SyncPlan) FullLength() int {
	return s.localLength() + s.remoteLength() + s.conflictLength()
}

func (s SyncPlan) IsAnyLocalChange() bool {
	return s.localLength() > 0
}

func (s SyncPlan) IsAnyRemoteChange() bool {
	return s.remoteLength() > 0
}

func (s SyncPlan) IsAnyConflict() bool {
	return s.conflictLength() > 0
}

func (s SyncPlan) LocalLength() int {
	return s.localLength()
}

func (s SyncPlan) RemoteLength() int {
	return s.remoteLength()
}

func (s SyncPlan) ChangeLength() int {
	return s.localLength() + s.remoteLength()
}

func (s SyncPlan) ConflictLength() int {
	return s.conflictLength()
}

func (s SyncPlan) conflictLength() int {
	return len(s.Conflicts)
}

func (s SyncPlan) localLength() int {
	return len(s.LocalChanges)
}

func (s SyncPlan) remoteLength() int {
	return len(s.RemoteChanges)
}

// При создании плана сортируем по путям лексикографически
func NewSyncPlan(localChanges []LocalChange, remoteChanges []RemoteChange, conflicts []Conflict) SyncPlan {
	sort.Slice(localChanges, func(i, j int) bool {
		return localChanges[i].ToSyncChange().SortPath().String() < localChanges[j].ToSyncChange().SortPath().String()
	})

	sort.Slice(remoteChanges, func(i, j int) bool {
		return remoteChanges[i].ToSyncChange().SortPath().String() < remoteChanges[j].ToSyncChange().SortPath().String()
	})

	sort.Slice(conflicts, func(i, j int) bool {
		return conflicts[i].LocalRelativePath.String() < conflicts[j].LocalRelativePath.String()
	})

	return SyncPlan{
		LocalChanges:  localChanges,
		RemoteChanges: remoteChanges,
		Conflicts:     conflicts,
	}
}
