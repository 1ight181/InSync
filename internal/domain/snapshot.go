package domain

type Snapshot struct {
	RootName RootName
	UnixTime uint64
	Files    []FileEntry
}
