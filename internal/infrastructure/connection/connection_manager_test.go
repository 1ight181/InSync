package connection

import (
	"io"
	"log/slog"
	"testing"

	"insync/internal/domain"
	clt "insync/internal/infrastructure/client"
	"insync/internal/interfaces"

	"github.com/stretchr/testify/require"
)

type stubClientHolder struct{}

func (s *stubClientHolder) SetClient(client interfaces.IClient) {}
func (s *stubClientHolder) CurrentClient() (interfaces.IClient, error) {
	return nil, nil
}
func (s *stubClientHolder) ReleaseClient() {}

func TestNewConnectionManager_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	manager, err := NewConnectionManager(ConnectionManagerOptions{
		BaseGrpcConf:      &clt.GrpcConf{},
		ClientHolder:      &stubClientHolder{},
		ServerServiceType: "service-type",
		ServerNamePrefix:  "name-prefix",
		GrpcClientLogger:  logger,
		Logger:            logger,
	})

	require.NoError(t, err)
	require.NotNil(t, manager)
}

func TestNewConnectionManager_InvalidOpts_ReturnsError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	manager, err := NewConnectionManager(ConnectionManagerOptions{
		BaseGrpcConf:     &clt.GrpcConf{},
		ClientHolder:     &stubClientHolder{},
		ServerNamePrefix: "name-prefix",
		GrpcClientLogger: logger,
		Logger:           logger,
	})

	require.ErrorIs(t, err, ErrInvalidOpts)
	require.Nil(t, manager)
}

func TestCurrentNodeName_NotConnected_ReturnsError(t *testing.T) {
	manager := &ConnectionManager{}

	_, err := manager.CurrentNodeName()
	require.ErrorIs(t, err, domain.ErrNotConnected)
}

func TestResolveServerName_ReplacesServiceNamePrefix(t *testing.T) {
	manager := &ConnectionManager{
		serverServiceType: "serverprefix",
		serverNamePrefix:  "nameprefix",
	}

	result := manager.resolveServerName(domain.NodeName("deviceid.serverprefix.domain."))
	require.Equal(t, "deviceid.nameprefix.domain", result)
}
