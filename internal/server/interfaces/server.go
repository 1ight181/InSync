package interfaces

type Server interface {
	Start() error
	Stop()
}
