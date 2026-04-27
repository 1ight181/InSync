package connection

import (
	"context"
	"errors"
	"insync/internal/domain"
	clt "insync/internal/infrastructure/client"
	"insync/internal/interfaces"
	"log/slog"
	"strings"
)

type ConnectionManager struct {
	baseGrpcConf *clt.GrpcConf

	grpcClientLogger *slog.Logger

	logger    *slog.Logger
	loggerCtx context.Context

	clientConnector IClientConnector
	clientHolder    IClientHolder

	serverServiceType string
	serverNamePrefix  string

	currentNodeName domain.NodeName
}

type ConnectionManagerOptions struct {
	BaseGrpcConf *clt.GrpcConf

	ClientConnector IClientConnector
	ClientHolder    IClientHolder

	ServerServiceType string
	ServerNamePrefix  string

	GrpcClientLogger *slog.Logger

	Logger *slog.Logger
}

var (
	ErrInvalidOpts = errors.New("Все поля ConnectionManagerOptions должны быть заполнены")
)

func NewConnectionManager(opts ConnectionManagerOptions) (*ConnectionManager, error) {
	if opts.BaseGrpcConf == nil ||
		opts.ClientHolder == nil ||
		opts.ServerServiceType == "" ||
		opts.ServerNamePrefix == "" ||
		opts.GrpcClientLogger == nil ||
		opts.Logger == nil {
		return nil, ErrInvalidOpts
	}
	return &ConnectionManager{
		baseGrpcConf: opts.BaseGrpcConf,

		clientHolder: opts.ClientHolder,

		serverServiceType: opts.ServerServiceType,
		serverNamePrefix:  opts.ServerNamePrefix,

		grpcClientLogger: opts.GrpcClientLogger,
		logger:           opts.Logger,
	}, nil
}

func (c *ConnectionManager) CurrentClient() (interfaces.IClient, error) {
	return c.clientHolder.CurrentClient()
}

func (c *ConnectionManager) CurrentNodeName() (domain.NodeName, error) {
	if c.currentNodeName == "" {
		return "", domain.ErrNotConnected
	}

	return c.currentNodeName, nil
}

func (c *ConnectionManager) ConnectToNode(nodeName domain.NodeName) error {
	grpcConf := c.baseGrpcConf
	grpcConf.ServerAddress = nodeName.String()
	// Причина почему используется формат deviceid.serverprefix.domain., а не
	// deviceid._service._proto.domain. в том, что wildcard не поддерживается для адресов с нижним подчеркиванием
	grpcConf.ServerName = c.resolveServerName(nodeName)

	grpcClientOpts := clt.GrpcClientOptions{
		Conf:   grpcConf,
		Logger: c.grpcClientLogger,
	}

	newClient, err := clt.NewGrpcClient(grpcClientOpts)
	if err != nil {
		return err
	}

	if err := newClient.Connect(); err != nil {
		return err
	}

	oldClient := c.clientConnector
	if oldClient != nil {
		if err := oldClient.Close(); err != nil {
			return err
		}
	}

	c.clientConnector = newClient
	c.clientHolder.SetClient(newClient)
	c.currentNodeName = nodeName

	c.logger.LogAttrs(
		c.loggerCtx,
		slog.LevelDebug,
		"Успешное подключение к узлу",
		slog.String("nodeName", nodeName.String()),
	)

	return nil
}

func (c *ConnectionManager) Close() error {
	c.clientHolder.ReleaseClient()

	if c.clientConnector != nil {
		return c.clientConnector.Close()
	}

	return nil
}

func (c *ConnectionManager) resolveServerName(nodeName domain.NodeName) string {
	serverName := strings.Replace(nodeName.String(), c.serverServiceType, c.serverNamePrefix, 1)
	serverNameWithoutDomainDot, _ := strings.CutSuffix(serverName, ".")

	return serverNameWithoutDomainDot
}
