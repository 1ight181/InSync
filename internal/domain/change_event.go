package domain

type ChangeEvent struct {
	Change SyncChange
	Err    error
}
