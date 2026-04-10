package holder

import (
	"errors"
	"insync/internal/interfaces"
)

var ErrNoClient = errors.New("Клиент не выбран")

type ClientHolder struct {
	client interfaces.IClient
}

func NewClientHolder() *ClientHolder {
	return &ClientHolder{}
}

func (cs *ClientHolder) CurrentClient() (interfaces.IClient, error) {
	if cs.client == nil {
		return nil, ErrNoClient
	}
	return cs.client, nil
}

func (cs *ClientHolder) SetClient(client interfaces.IClient) {
	cs.client = client
}

func (cs *ClientHolder) ReleaseClient() {
	cs.client = nil
}
