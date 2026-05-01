package app

import (
	"fmt"
	aliasprov "insync/internal/infrastructure/alias"
	"insync/internal/infrastructure/base"
	"insync/internal/infrastructure/config/models"
	conn "insync/internal/infrastructure/connection"
	"insync/internal/infrastructure/deviceid"
	"insync/internal/infrastructure/filemanager"
	local "insync/internal/infrastructure/local"
	planner "insync/internal/infrastructure/planner"
	planres "insync/internal/infrastructure/planresolver"
	remote "insync/internal/infrastructure/remote"
	"insync/internal/infrastructure/root"
	syncer "insync/internal/infrastructure/syncer"
	changeappl "insync/internal/infrastructure/syncer/applier"
	persister "insync/internal/infrastructure/syncer/persister"
	cli "insync/internal/presentation/cli"
	aliasrepo "insync/internal/repository/sqlite/alias"
	aliasusecase "insync/internal/usecase/alias"
	baseusecase "insync/internal/usecase/base"
	connusecase "insync/internal/usecase/connect"
	initusecase "insync/internal/usecase/init"
	nodeusecase "insync/internal/usecase/node"
	rootusecase "insync/internal/usecase/root"
	scanusecase "insync/internal/usecase/scan"
	syncusecase "insync/internal/usecase/sync"
	"log/slog"

	"gorm.io/gorm"
)

func createCLi(
	cliLogger *slog.Logger,
	mDnsBrowserLogger *slog.Logger,
	mDnsBrowserConfig models.MDnsBrowserConfig,
	baseSnapshotManager *base.BaseSnapshotManager,
	fileManager *filemanager.FileManager,
	connectionManager *conn.ConnectionManager,
	deviceIdProvider *deviceid.DeviceIdProvider,
	localSnapshotProvider *local.LocalSnapshotProvider,
	rootResolver *root.RootResolver,
	db *gorm.DB,
	baseUseCase *baseusecase.BaseSnapshotUseCase,
) (*cli.Cli, error) {
	connectUseCase, err := connusecase.NewConnectUseCase(connectionManager)
	if err != nil {
		return nil, err
	}

	aliasRepository, err := aliasrepo.NewAliasRepository(db)
	if err != nil {
		return nil, err
	}

	aliasProvider, err := aliasprov.NewAliasProvider(aliasRepository)
	if err != nil {
		return nil, err
	}

	mDnsNodeNamesBrowser, err := createMDnsNodeNamesBrowser(
		mDnsBrowserLogger,
		mDnsBrowserConfig,
	)
	if err != nil {
		return nil, err
	}

	nodeUseCaseOpts := nodeusecase.NodeUseCaseOptions{
		NodeNamesBrowser:      mDnsNodeNamesBrowser,
		LocalDeviceIdResolver: deviceIdProvider,
	}

	nodeUseCase, err := nodeusecase.NewNodeUseCase(nodeUseCaseOpts)
	if err != nil {
		return nil, err
	}

	remoteSnapshotProvider, err := remote.NewRemoteSnapshotProvider(connectionManager)
	if err != nil {
		return nil, err
	}

	planner := planner.NewChangesPlannerWithTreeSkip()

	planCache := planres.NewSyncPlanCache()

	planResolverOpts := planres.PlanResolverOptions{
		BaseSnapshotProvider:   baseSnapshotManager,
		RemoteSnapshotProvider: remoteSnapshotProvider,
		LocalSnapshotProvider:  localSnapshotProvider,

		ChangesPlanner: planner,
		SyncPlanCache:  planCache,
	}

	planResolver, err := planres.NewPlanResolver(planResolverOpts)
	if err != nil {
		return nil, err
	}

	scanUseCase, err := scanusecase.NewScanUseCase(planResolver)
	if err != nil {
		return nil, err
	}

	changeApplierOpts := changeappl.ChangeApplierOptions{
		FileManager:   fileManager,
		ClientFactory: connectionManager,
	}

	changeApplier, err := changeappl.NewChangeApplier(changeApplierOpts)
	if err != nil {
		return nil, err
	}

	conflictResolver := syncer.NewConflictResolver()

	postSyncBaseSnapshotPersisterOpts := persister.PostSyncBaseSnapshotPersisterOptions{
		BaseSnapshotManager:   baseSnapshotManager,
		LocalSnapshotProvider: localSnapshotProvider,
		ClientFactory:         connectionManager,
	}

	postSyncBaseSnapshotPersister, err := persister.NewPostSyncBaseSnapshotPersister(postSyncBaseSnapshotPersisterOpts)
	if err != nil {
		return nil, err
	}

	syncerOpts := syncer.SyncerOptions{
		ChangeApplier:                 changeApplier,
		ConflictResolver:              conflictResolver,
		PostSyncBaseSnapshotPersister: postSyncBaseSnapshotPersister,
	}

	syncer, err := syncer.NewSyncer(syncerOpts)
	if err != nil {
		return nil, err
	}

	syncusecaseOpts := syncusecase.SyncUseCaseOptions{
		SyncPlanResolver: planResolver,
		Syncer:           syncer,
	}

	syncUseCase, err := syncusecase.NewSyncUseCase(syncusecaseOpts)
	if err != nil {
		return nil, err
	}

	initUseCaseOpts := initusecase.InitUseCaseOptions{
		BaseSnapshotManager:   baseSnapshotManager,
		LocalSnapshotProvider: localSnapshotProvider,
	}

	initUseCase, err := initusecase.NewInitUseCase(initUseCaseOpts)
	if err != nil {
		return nil, err
	}

	aliasUseCaseOpts := aliasusecase.AliasUseCaseOptions{
		AliasProvider: aliasProvider,
	}

	aliasUseCase, err := aliasusecase.NewAliasUseCase(aliasUseCaseOpts)
	if err != nil {
		return nil, err
	}

	rootUseCase, err := rootusecase.NewRootUseCase(rootResolver)
	if err != nil {
		return nil, err
	}

	cliInstanceOpts := cli.CliOptions{
		SyncUseCase:    syncUseCase,
		ScanUseCase:    scanUseCase,
		NodeUseCase:    nodeUseCase,
		ConnectUseCase: connectUseCase,
		RootUseCase:    rootUseCase,
		InitUseCase:    initUseCase,
		AliasUseCase:   aliasUseCase,
		BaseUseCase:    baseUseCase,

		Logger: cliLogger,
	}

	cliInstance, err := cli.NewCli(cliInstanceOpts)
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать Cli: %v", err))
	}

	return cliInstance, nil
}
