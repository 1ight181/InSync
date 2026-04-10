package cli

type IConnectUseCase interface {
	ConnectToNode(nodeName string) error
	CurrentNode() (nodeName string, err error)
	ShowNodes() (nodeNamesChan chan string, err error)
}
