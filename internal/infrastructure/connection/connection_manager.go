package connection

import (
	"context"
	"errors"
	"fmt"
	"insync/internal/domain"
	clt "insync/internal/infrastructure/client"
	"insync/internal/interfaces"
	"log/slog"
)

type ConnectionManager struct {
	baseGrpcConf *clt.GrpcConf

	grpcClientLogger *slog.Logger

	logger    *slog.Logger
	loggerCtx context.Context

	clientConnector IClientConnector
	clientHolder    IClientHolder

	currentNodeName domain.NodeName

	mDnsServerServiceType         string
	mDnsServerDomain              string
	mDnsServerInstanceNamePostfix string
}

type ConnectionManagerOptions struct {
	BaseGrpcConf *clt.GrpcConf

	ClientConnector IClientConnector
	ClientHolder    IClientHolder

	GrpcClientLogger *slog.Logger

	Logger *slog.Logger

	MDnsServerServiceType         string
	MDnsServerDomain              string
	MDnsServerInstanceNamePostfix string
}

var (
	ErrInvalidOpts = errors.New("Все поля ConnectionManagerOptions должны быть заполнены")
)

func NewConnectionManager(opts ConnectionManagerOptions) (*ConnectionManager, error) {
	if opts.BaseGrpcConf == nil ||
		opts.ClientHolder == nil ||
		opts.GrpcClientLogger == nil ||
		opts.MDnsServerServiceType == "" ||
		opts.MDnsServerDomain == "" ||
		opts.MDnsServerInstanceNamePostfix == "" ||
		opts.Logger == nil {
		return nil, ErrInvalidOpts
	}
	return &ConnectionManager{
		baseGrpcConf: opts.BaseGrpcConf,

		clientHolder: opts.ClientHolder,

		grpcClientLogger: opts.GrpcClientLogger,
		logger:           opts.Logger,

		mDnsServerServiceType:         opts.MDnsServerServiceType,
		mDnsServerDomain:              opts.MDnsServerDomain,
		mDnsServerInstanceNamePostfix: opts.MDnsServerInstanceNamePostfix,
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
	mDnsUrl := c.resolveMDnsUrl(nodeName)
	serverName := c.resolveServerName(nodeName)

	grpcConf := c.baseGrpcConf
	grpcConf.ServerAddress = mDnsUrl
	grpcConf.ServerName = serverName

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
		slog.String("mDnsUrl", mDnsUrl),
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

func (c *ConnectionManager) resolveMDnsUrl(nodeName domain.NodeName) string {
	return fmt.Sprintf("%s.%s.%s.%s", c.mDnsServerServiceType, nodeName, c.mDnsServerInstanceNamePostfix, c.mDnsServerDomain)
}

func (c *ConnectionManager) resolveServerName(nodeName domain.NodeName) string {
	return fmt.Sprintf("%s.%s.%s", nodeName, c.mDnsServerInstanceNamePostfix, c.mDnsServerDomain)
}
