package domain

type Decision int

const (
	LocalWin Decision = iota
	RemoteWin
)

func (d Decision) Equal(other Decision) bool {
	return d == other
}
