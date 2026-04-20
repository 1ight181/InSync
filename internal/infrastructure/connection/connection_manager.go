package connection

import (
	"context"
	"insync/internal/domain"
	clt "insync/internal/infrastructure/client"
	"insync/internal/interfaces"
	"log/slog"
)

type ConnectionManager struct {
	mdnsUrlResolver IMDnsUrlResolver
	baseGrpcConf    *clt.GrpcConf

	grpcClientLogger *slog.Logger

	logger    *slog.Logger
	loggerCtx context.Context

	clientConnector IClientConnector
	clientHolder    IClientHolder

	currentNodeName domain.NodeName
}

type ConnectionManagerOptions struct {
	MDnsUrlResolver IMDnsUrlResolver
	BaseGrpcConf    *clt.GrpcConf

	GrpcClientLogger *slog.Logger

	Logger *slog.Logger
}

func NewConnectionManager(opts ConnectionManagerOptions) *ConnectionManager {
	if opts.BaseGrpcConf == nil ||
		opts.MDnsUrlResolver == nil ||
		opts.GrpcClientLogger == nil ||
		opts.Logger == nil {
		panic("Все поля ConnectionManagerOptions должны быть заполнены")
	}
	return &ConnectionManager{
		mdnsUrlResolver: opts.MDnsUrlResolver,
		baseGrpcConf:    opts.BaseGrpcConf,

		grpcClientLogger: opts.GrpcClientLogger,
		logger:           opts.Logger,
	}
}

func (c *ConnectionManager) CurrentClient() interfaces.IClient {
	return c.clientHolder.CurrentClient()
}

func (c *ConnectionManager) CurrentNodeName() domain.NodeName {
	return c.currentNodeName
}

func (c *ConnectionManager) ConnectToNode(nodeName domain.NodeName) error {
	mDnsUrl := c.mdnsUrlResolver.Resolve(nodeName)

	grpcConf := c.baseGrpcConf
	grpcConf.ServerAddress = mDnsUrl

	grpcClientOpts := clt.GrpcClientOptions{
		Conf:   grpcConf,
		Ctx:    nil,
		Logger: nil,
	}

	newClient := clt.NewGrpcClient(grpcClientOpts)

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
