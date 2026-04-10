package connect

type IConnectionManager interface {
	ConnectToNode(nodeName string) error
	CurrentNodeName() (nodeName string, err error)
}
