package domain

type Decision int

const (
	LocalWin Decision = iota
	RemoteWin
	Skip
)

func (d Decision) Equal(other Decision) bool {
	return d == other
}
