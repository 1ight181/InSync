package connection

type IClientConnector interface {
	Connect() error
	Close() error
}
