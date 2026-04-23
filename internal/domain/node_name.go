package domain

import "errors"

type NodeName string

var ErrEmptyNodeName = errors.New("NodeName не может быть пустым")

func NewNodeName(name string) (NodeName, error) {
	if name == "" {
		return "", ErrEmptyNodeName
	}
	return NodeName(name), nil
}

func (nodeName NodeName) String() string {
	return string(nodeName)
}

var (
	ErrNotConnected = errors.New("Узел не подключен")
)
