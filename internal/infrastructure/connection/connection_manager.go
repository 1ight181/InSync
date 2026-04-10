package connection

import (
	"context"
	clt "insync/internal/infrastructure/client"
	"insync/internal/interfaces"
	"log/slog"
)

type ConnectionManager struct {
	nodeNameResolver INodeNameResolver
	baseGrpcConf     *clt.GrpcConf

	ctx context.Context

	grpcClientLogger *slog.Logger

	logger    *slog.Logger
	loggerCtx context.Context

	clientConnector IClientConnector
	client          interfaces.IClient
	currentNodeName string
}

type ConnectionManagerOptions struct {
	NodeNameResolver INodeNameResolver
	BaseGrpcConf     *clt.GrpcConf

	Ctx context.Context

	GrpcClientLogger *slog.Logger

	Logger *slog.Logger
}

func NewConnectionManager(opts ConnectionManagerOptions) *ConnectionManager {
	if opts.BaseGrpcConf == nil ||
		opts.NodeNameResolver == nil ||
		opts.Ctx == nil ||
		opts.GrpcClientLogger == nil ||
		opts.Logger == nil {
		panic("Все поля ConnectionManagerOptions должны быть заполнены")
	}
	return &ConnectionManager{
		nodeNameResolver: opts.NodeNameResolver,
		baseGrpcConf:     opts.BaseGrpcConf,

		ctx: opts.Ctx,

		grpcClientLogger: opts.GrpcClientLogger,
		logger:           opts.Logger,
	}
}

func (c *ConnectionManager) CurrentClient() interfaces.IClient {
	return c.client
}

func (c *ConnectionManager) CurrentNodeName() string {
	return c.currentNodeName
}

func (c *ConnectionManager) ConnectToNode(nodeName string) error {
	mdnsUrl, err := c.nodeNameResolver.ResolveToMDnsUrl(nodeName)
	if err != nil {
		return err
	}

	grpcConf := c.baseGrpcConf
	grpcConf.ServerAddress = mdnsUrl

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
	c.client = newClient
	c.currentNodeName = nodeName

	return nil
}
