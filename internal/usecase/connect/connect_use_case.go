package connect

type ConnectUseCase struct {
	nodeNamesBrowser  INodeNamesBrowser
	connectionManager IConnectionManager
}

type ConnectUseCaseOptions struct {
	NodeNamesBrowser INodeNamesBrowser
}

func NewConnectUseCase(opts ConnectUseCaseOptions) *ConnectUseCase {
	if opts.NodeNamesBrowser == nil {
		panic("Все поля ConnectUseCaseOptions должны быть заполнены")
	}
	return &ConnectUseCase{
		nodeNamesBrowser: opts.NodeNamesBrowser,
	}
}

func (c *ConnectUseCase) ShowNodes() (nodeNamesChan chan string, err error) {
	return c.nodeNamesBrowser.BrowseNodeNames()
}

func (c *ConnectUseCase) CurrentNode() (nodeName string, err error) {
	return c.connectionManager.CurrentNodeName()
}

func (c *ConnectUseCase) ConnectToNode(nodeName string) error {
	return c.connectionManager.ConnectToNode(nodeName)
}
