package connection

type INodeNameResolver interface {
	Resolve(nodeName string) (string, error)
}
