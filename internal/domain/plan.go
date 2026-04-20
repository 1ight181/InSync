package domain

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

func (s SyncPlan) localLength() int {
	return len(s.LocalChanges)
}

func (s SyncPlan) remoteLength() int {
	return len(s.RemoteChanges)
}

func (s SyncPlan) conflictLength() int {
	return len(s.Conflicts)
}
