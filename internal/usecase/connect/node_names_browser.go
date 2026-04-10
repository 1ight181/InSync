package connect

type INodeNamesBrowser interface {
	BrowseNodeNames() (nodeNamesChan chan string, err error)
}
