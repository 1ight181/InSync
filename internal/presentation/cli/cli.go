package cli

import (
	"context"
	"errors"
	"fmt"
	"insync/internal/domain"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unsafe"

	cobraprompt "github.com/1ight181/comptplus-ctrl-c"
	prompt "github.com/1ight181/go-prompt-ctrl-c"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows"

	promptui "github.com/manifoldco/promptui"
)

const (
	nodesCmdName       = "nodes"
	currentNodeCmdName = "current-node"
	addRootCmdName     = "add-root"
	removeRootCmdName  = "remove-root"
	getRootsCmdName    = "roots"
	connectCmdName     = "connect"
	dryRunCmdName      = "dry-run"
	syncCmdName        = "sync"
	initCmdName        = "init"
	setAliasCmdName    = "set-alias"
	removeAliasCmdName = "remove-alias"
)

const (
	conflictTypeLocalDeletedRemoteModified    = "Файл удален на локальном узле и модифицирован на удалённом"
	conflictTypeRemoteDeletedLocalModified    = "Файл удален на удалённом узле и модифицирован на локальном"
	conflictTypeLocalMovedRemoteMoved         = "Файл перемещен и на локальном узле, и на удалённом"
	conflictTypeLocalRenamedRemoteRenamed     = "Файл переименован и на локальном узле, и на удалённом"
	conflictTypeBothModifiedAtSameTime        = "Обе стороны модифицировали один и тот же файл в одно и то же время"
	conflictTypeBothCreatedAtSamePathConflict = "Обе стороны создали файл по одному и тому же пути, но содержимое отличается"
	conflictTypeUnknown                       = "Неизвестный тип конфликта"
)

const (
	decisionLocalWin  = "Применить локальное изменение"
	decisionRemoteWin = "Применить удалённое изменение"
	decisionSkip      = "Пропустить конфликт"
	decisionUnknown   = "Неизвестное решение"
)

const (
	shouldUseCacheFlagName        = "should_use_cache"
	shouldDeleteSnapshotsFlagName = "should_delete_snapshots"
)

const (
	changeTypeCreate  = "CREATE"
	changeTypeDelete  = "DELETE"
	changeTypeRename  = "RENAME"
	changeTypeMove    = "MOVE"
	changeTypeModify  = "MODIFY"
	changeTypeUnknown = "UNKNOWN"
)

type Cli struct {
	syncUseCase    ISyncUseCase
	scanUseCase    IScanUseCase
	nodeUseCase    INodeUseCase
	connectUseCase IConnectUseCase
	rootUseCase    IRootUseCase
	initUseCase    IInitUseCase
	aliasUseCase   IAliasUseCase
	baseUseCase    IBaseUseCase

	logger    *slog.Logger
	loggerCtx context.Context

	shouldUseCache        *bool
	shouldDeleteSnapshots *bool
	nodeNameCache         []domain.NodeName
	exitCtxCancel         context.CancelFunc
}

type CliOptions struct {
	SyncUseCase    ISyncUseCase
	ScanUseCase    IScanUseCase
	NodeUseCase    INodeUseCase
	ConnectUseCase IConnectUseCase
	RootUseCase    IRootUseCase
	InitUseCase    IInitUseCase
	AliasUseCase   IAliasUseCase
	BaseUseCase    IBaseUseCase
	Logger         *slog.Logger
}

var (
	ErrInvalidCliOptions   = errors.New("Все поля CliOptions должны быть заполнены")
	ErrInvalidBaseProvider = errors.New("Базовый провайдер должен быть remote или local")
)

func NewCli(opts CliOptions) (*Cli, error) {
	if opts.SyncUseCase == nil ||
		opts.ScanUseCase == nil ||
		opts.NodeUseCase == nil ||
		opts.ConnectUseCase == nil ||
		opts.RootUseCase == nil ||
		opts.InitUseCase == nil ||
		opts.AliasUseCase == nil ||
		opts.Logger == nil ||
		opts.BaseUseCase == nil {
		return nil, ErrInvalidCliOptions
	}
	loggerCtx := context.Background()
	return &Cli{
		syncUseCase:    opts.SyncUseCase,
		scanUseCase:    opts.ScanUseCase,
		nodeUseCase:    opts.NodeUseCase,
		connectUseCase: opts.ConnectUseCase,
		rootUseCase:    opts.RootUseCase,
		initUseCase:    opts.InitUseCase,
		aliasUseCase:   opts.AliasUseCase,
		baseUseCase:    opts.BaseUseCase,

		logger:    opts.Logger,
		loggerCtx: loggerCtx,
	}, nil
}

func (c *Cli) Start() {
	rootCmd := c.createRootCmd()

	initCmd := c.createInitCmd()

	nodesCmd := c.createNodesCmd()
	connectCmd := c.createConnectCmd()
	currentNodeCmd := c.createCurrentNodeCmd()

	setAliasCmd := c.createSetAliasCmd()
	removeAliasCmd := c.createRemoveAliasCmd()

	addRootCmd := c.createAddRootCmd()
	removeRootCmd := c.createRemoveRootCmd()
	getRootsCmd := c.createRootsCmd()

	dryRunCmd := c.createDryRunCmd()
	syncCmd := c.createSyncCmd()

	rootCmd.AddCommand(nodesCmd)
	rootCmd.AddCommand(currentNodeCmd)

	rootCmd.AddCommand(initCmd)

	rootCmd.AddCommand(addRootCmd)
	rootCmd.AddCommand(removeRootCmd)
	rootCmd.AddCommand(getRootsCmd)

	rootCmd.AddCommand(setAliasCmd)
	rootCmd.AddCommand(removeAliasCmd)

	rootCmd.AddCommand(connectCmd)

	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(dryRunCmd)

	cobraPrompt := cobraprompt.CobraPrompt{
		RootCmd:                 rootCmd,
		ShowHelpCommandAndFlags: true,
		GoPromptOptions: []prompt.Option{
			prompt.WithTitle("InSync 0.1"),
			prompt.WithMaxSuggestion(5),
			prompt.WithPrefix("insync> "),
			prompt.WithInputTextColor(prompt.DarkGreen),
			prompt.WithSelectedSuggestionTextColor(prompt.Green),
			prompt.WithSuggestionTextColor(prompt.DarkGreen),
			prompt.WithInterruptCallback(c.interruptCallback),
		},
		DynamicSuggestionsFunc: c.suggestionFunc,
		OnErrorFunc: func(err error) {
			if strings.Contains(err.Error(), "unknown command") {
				fmt.Print("Неизвестная команда!\n")
			} else {
				fmt.Printf("%s\n", err)
			}
		},
		InArgsParser: c.parseWindowsCommandLine,
	}
	cobraPrompt.Run()
}

func (c *Cli) interruptCallback(code int) {
}

func (c *Cli) parseWindowsCommandLine(commandLine string) []string {
	commandLinePointer, err := syscall.UTF16PtrFromString(commandLine)
	if err != nil {
		return strings.Fields(commandLine)
	}

	var argumentCount int32

	argumentVector, err := windows.CommandLineToArgv(commandLinePointer, &argumentCount)
	if err != nil {
		return strings.Fields(commandLine)
	}
	defer windows.LocalFree(windows.Handle(uintptr(unsafe.Pointer(argumentVector))))

	arguments := make([]string, 0, argumentCount)

	for index := int32(0); index < argumentCount; index++ {
		argument := syscall.UTF16ToString(argumentVector[index][:])
		arguments = append(arguments, argument)
	}

	return arguments
}

func (c *Cli) createRootCmd() *cobra.Command {
	return &cobra.Command{
		SilenceUsage:  true,
		SilenceErrors: true,
		Short:         "InSync - инструмент для синхронизации файлов между различными хранилищами",
	}
}

func (c *Cli) createSetAliasCmd() *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s node-name alias-name", setAliasCmdName),
		Short: "Установить псевдоним для nodeName",
		RunE:  c.setAliasCmd,
		Long: `Установить псевдоним для nodeName
		node-name - имя узла
		alias-name - имя псевдонима`,
		Args: cobra.ExactArgs(2),
		Annotations: map[string]string{
			cobraprompt.DynamicSuggestionsAnnotation: setAliasCmdName,
		},
	}
}

func (c *Cli) createRemoveAliasCmd() *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s alias-name", removeAliasCmdName),
		Short: "Удалить псевдоним",
		RunE:  c.removeAliasCmd,
		Long: `Удалить псевдоним. 
		alias-name - имя псевдонима`,
		Args: cobra.ExactArgs(1),
		Annotations: map[string]string{
			cobraprompt.DynamicSuggestionsAnnotation: removeAliasCmdName,
		},
	}
}

func (c *Cli) createInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s base_snapshot_provider root_name", initCmdName),
		Short: "Инициализация",
		Long: `Инициализация. Создает базовый снимок. Если вызывается повторно, то перезаписывает текущий базовый снимок.
		root_name - имя корневого каталога
		`,
		Args: cobra.ExactArgs(1),
		RunE: c.initCmd,
		Annotations: map[string]string{
			cobraprompt.DynamicSuggestionsAnnotation: initCmdName,
		},
	}
}

func (c *Cli) createNodesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   nodesCmdName,
		Short: "Отобразить доступные для синхронизации узлы",
		Args:  cobra.NoArgs,
		RunE:  c.nodesCmd,
	}
}

func (c *Cli) createCurrentNodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   currentNodeCmdName,
		Short: "Отобразить текущий узел",
		Args:  cobra.NoArgs,
		RunE:  c.currentNodeCmd,
	}
}

func (c *Cli) createAddRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s root-name root-path", addRootCmdName),
		Short: "Добавить корневой каталог",
		Long: `Добавить корневой каталог. 
		root-name - имя корневого каталога
		root-path - путь к корневому каталогу`,
		RunE: c.addRootCmd,
		Args: cobra.ExactArgs(2),
	}
	return cmd
}

func (c *Cli) createRemoveRootCmd() *cobra.Command {
	removeRootCmd := &cobra.Command{
		Use:   fmt.Sprintf("%s root-name", removeRootCmdName),
		Short: "Удалить корневой каталог",
		Long: `Удалить корневой каталог. 
		root-name - имя корневого каталога`,
		RunE: c.removeRootCmd,
		Args: cobra.ExactArgs(1),
		Annotations: map[string]string{
			cobraprompt.DynamicSuggestionsAnnotation: removeRootCmdName,
		},
	}

	shouldDeleteSnapshotsPtr := removeRootCmd.Flags().Bool(
		shouldDeleteSnapshotsFlagName, false,
		"Указывает стоит ли удалять связанные снимки при удалении корневого каталога",
	)

	c.shouldDeleteSnapshots = shouldDeleteSnapshotsPtr

	return removeRootCmd
}

func (c *Cli) createRootsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s", getRootsCmdName),
		Short: "Получить список корневых каталогов",
		Args:  cobra.NoArgs,
		Run:   c.rootsCmd,
	}
}

func (c *Cli) createConnectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s node-name", connectCmdName),
		Short: "Подключиться к узлу",
		Long: `Подключиться к узлу. 
		node-name - имя узла, к которому будет осуществляться подключение`,
		RunE: c.connectCmd,
		Args: cobra.ExactArgs(1),
		Annotations: map[string]string{
			cobraprompt.DynamicSuggestionsAnnotation: connectCmdName,
		},
	}
}

func (c *Cli) createDryRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s root-name", dryRunCmdName),
		Short: "Показать, какие файлы/каталоги будут синхронизированы, без фактического выполнения синхронизации",
		Long: `Показать, какие файлы/каталоги будут синхронизированы, без фактического выполнения синхронизации. 
		root-name - имя корневого каталога`,
		RunE: c.dryRunCmd,
		Args: cobra.ExactArgs(1),
		Annotations: map[string]string{
			cobraprompt.DynamicSuggestionsAnnotation: dryRunCmdName,
		},
	}
}

func (c *Cli) createSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s root-name", syncCmdName),
		Short: "Выполнить синхронизацию",
		Long: `Выполнить синхронизацию. 
		root-name - имя корневого каталога`,
		RunE: c.syncCmd,
		Args: cobra.ExactArgs(1),
		Annotations: map[string]string{
			cobraprompt.DynamicSuggestionsAnnotation: syncCmdName,
		},
	}

	shouldUseCachePtr := cmd.Flags().Bool(
		shouldUseCacheFlagName,
		true,
		"Указывает стоит ли использовать изенения полученные с последнего скана или стоит просканировать еще раз",
	)

	c.shouldUseCache = shouldUseCachePtr

	return cmd
}

func (c *Cli) initCmd(cmd *cobra.Command, args []string) error {
	c.logger.Debug("Выполнение команды init")

	rootName, err := domain.NewRootName(args[0])
	if err != nil {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Не удалось создать корневой каталог", slog.String("error", err.Error()))
		fmt.Println("Не удалось создать корневой каталог")
		return err
	}

	cmdCtx := cmd.Context()
	interruptCtx, interruptCancel := signal.NotifyContext(cmdCtx, os.Interrupt)
	defer interruptCancel()

	if err := c.initUseCase.Init(interruptCtx, rootName); err != nil {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Не удалось выполнить init", slog.String("error", err.Error()))
		fmt.Println("Не удалось выполнить init")
		return err
	}

	fmt.Println("Базовый снимок создан")

	c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Выполнено init", slog.String("rootName", rootName.String()))

	return nil
}

func (c *Cli) nodesCmd(cmd *cobra.Command, args []string) error {
	c.logger.Debug("Выполнение команды nodes")

	cmdCtx := cmd.Context()
	interruptCtx, interruptCancel := signal.NotifyContext(cmdCtx, os.Interrupt)
	defer interruptCancel()

	nodeNamesChan, err := c.nodeUseCase.ShowNodeNames(interruptCtx)
	if err != nil {
		fmt.Print("Не удалось получить узлы\n")
		return err
	}

	fmt.Println("Доступные узлы (динамический список, нажмите Ctrl+C для завершения):")

	var newNodeNamesCache []domain.NodeName

	for i := 1; ; i++ {
		select {
		case nodeName, ok := <-nodeNamesChan:
			if ok {
				nodeNameCandidate := nodeName.String()
				if alias, ok := c.aliasUseCase.GetAliasByNode(nodeName); ok {
					nodeNameCandidate = alias
				}

				fmt.Printf("%d. %s\n", i, nodeNameCandidate)

				newNodeNamesCache = append(newNodeNamesCache, nodeName)
				c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Доступный узел", slog.String("name", nodeName.String()))
			}
		case <-interruptCtx.Done():
			c.nodeNameCache = newNodeNamesCache
			c.logger.Debug("Вызывано прерывание во время выполнения команды nodes")
			return nil
		}
	}
}

func (c *Cli) currentNodeCmd(cmd *cobra.Command, args []string) error {
	c.logger.Debug("Выполнение команды current-node")

	currentNode, err := c.connectUseCase.CurrentNodeName()
	if err != nil {
		if errors.Is(err, domain.ErrNotConnected) {
			c.logger.Debug("Нет подключенного узла")
			fmt.Println("Нет подключенного узла")
			return nil
		}

		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Не удалось получить текущий узел", slog.String("error", err.Error()))
		fmt.Println("Не удалось получить текущий узел")
		return err
	}
	c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Получен текущий узел", slog.String("node", currentNode.String()))
	fmt.Printf("Текущий подключенный узел: %s\n", currentNode)

	return nil
}

func (c *Cli) addRootCmd(cmd *cobra.Command, args []string) error {
	c.logger.Debug("Выполнение команды add-root")

	rootName, err := domain.NewRootName(args[0])
	if err != nil {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Не удалось создать корневой каталог", slog.String("error", err.Error()))
		fmt.Println("Не удалось создать корневой каталог")
		return err
	}

	rootRelativePath, err := domain.NewPath(args[1])
	if err != nil {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Не удалось создать корневой каталог", slog.String("error", err.Error()))
		fmt.Println("Не удалось создать корневой каталог")
		return err
	}

	ctx := cmd.Context()
	interruptCtx, interruptCancel := signal.NotifyContext(ctx, os.Interrupt)
	defer interruptCancel()

	if err := c.rootUseCase.AddRoot(interruptCtx, rootName, rootRelativePath); err != nil {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Не удалось добавить корневой каталог", slog.String("error", err.Error()))
		fmt.Println("Не удалось добавить корневой каталог")
		return err
	}

	c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Корневой каталог добавлен", slog.String("name", rootName.String()), slog.String("path", rootRelativePath.String()))
	fmt.Printf("Корневой каталог %s добавлен\n", rootName)
	return nil
}

func (c *Cli) removeRootCmd(cmd *cobra.Command, args []string) error {
	c.logger.Debug("Выполнение команды remove-root")

	rootName, err := domain.NewRootName(args[0])
	if err != nil {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Не удалось удалить корневой каталог", slog.String("error", err.Error()))
		fmt.Println("Не удалось создать удалить корневой каталог")
		return err
	}

	ctx := cmd.Context()
	interruptCtx, interruptCancel := signal.NotifyContext(ctx, os.Interrupt)
	defer interruptCancel()

	if err := c.baseUseCase.DeleteBaseSnapshots(interruptCtx, rootName); err != nil {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Не удалось удалить связанные снимки", slog.String("error", err.Error()))
		fmt.Println("Не удалось удалить связанные снимки")
		return err
	}

	if err := c.rootUseCase.RemoveRoot(interruptCtx, rootName); err != nil {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Не удалось удалить корневой каталог", slog.String("error", err.Error()))
		fmt.Println("Не удалось удалить корневой каталог")
		return err
	}

	c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Корневой каталог удален", slog.String("name", rootName.String()))
	fmt.Printf("Корневой каталог %s удален\n", rootName)

	return nil
}

func (c *Cli) rootsCmd(cmd *cobra.Command, args []string) {
	c.logger.Debug("Выполнение команды roots")

	roots, err := c.rootUseCase.GetRoots()
	if err != nil {
		if errors.Is(err, domain.ErrNoRoots) {
			c.logger.Debug("Нет корневых каталогов")
			fmt.Println("Нет корневых каталогов")
			return
		}
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Не удалось получить корневые каталоги", slog.String("error", err.Error()))
		fmt.Println("Не удалось получить корневые каталоги")
		return
	}

	c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Получен список корневых каталогов", slog.Int("count", len(roots)))
	fmt.Println("Список корневых каталогов:")
	var i int
	for root, path := range roots {
		i++
		fmt.Printf("%d. %s = %s\n", i, root, path)
	}
}
func (c *Cli) setAliasCmd(cmd *cobra.Command, args []string) error {
	c.logger.Debug("Выполнение команды set-alias")

	nodeName, err := domain.NewNodeName(args[0])
	if err != nil {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Не удалось установить псевдоним", slog.String("error", err.Error()))
		fmt.Println("Не удалось установить псевдоним")
		return err
	}

	alias := args[1]

	if alias == "" {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Не удалось установить псевдоним", slog.String("error", "псевдоним не может быть пустым"))
		fmt.Println("Не удалось установить псевдоним")
		return fmt.Errorf("псевдоним не может быть пустым")
	}

	ctx := cmd.Context()
	interruptCtx, interruptCancel := signal.NotifyContext(ctx, os.Interrupt)
	defer interruptCancel()

	if err := c.aliasUseCase.SetAlias(interruptCtx, alias, nodeName); err != nil {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Не удалось установить псевдоним", slog.String("error", err.Error()))
		fmt.Println("Не удалось установить псевдоним")
		return err
	}

	c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Псевдоним установлен", slog.String("alias", alias), slog.String("node", nodeName.String()))
	fmt.Printf("Псевдоним %s установлен\n", alias)
	return nil
}

func (c *Cli) removeAliasCmd(cmd *cobra.Command, args []string) error {
	c.logger.Debug("Выполнение команды remove-alias")

	alias := args[0]

	ctx := cmd.Context()
	interruptCtx, interruptCancel := signal.NotifyContext(ctx, os.Interrupt)
	defer interruptCancel()

	if err := c.aliasUseCase.RemoveAlias(interruptCtx, alias); err != nil {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Не удалось удалить псевдоним", slog.String("error", err.Error()))
		fmt.Println("Не удалось удалить псевдоним")
		return err
	}

	c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Псевдоним удален", slog.String("alias", alias))
	fmt.Printf("Псевдоним %s удален\n", alias)
	return nil
}

func (c *Cli) connectCmd(cmd *cobra.Command, args []string) error {
	c.logger.Debug("Выполнение команды connect")

	nodeName, err := domain.NewNodeName(args[0])
	if err != nil {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Не удалось подключиться к узлу", slog.String("error", err.Error()))
		fmt.Println("Не удалось подключиться к узлу")
		return err
	}

	var nodeNameToConnect domain.NodeName
	// Проверяем, не передан ли псевдоним в качестве аргумента
	if node, ok := c.aliasUseCase.GetNodeByAlias(nodeName.String()); ok {
		nodeNameToConnect = node
	}

	if nodeNameToConnect == "" {
		nodeNameToConnect = nodeName
	}

	c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Подключение к узлу", slog.String("node", nodeName.String()))
	fmt.Printf("Подключение к узлу %s\n", nodeName)

	if err := c.connectUseCase.ConnectToNode(nodeName); err != nil {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Не удалось подключиться к узлу", slog.String("error", err.Error()))
		return err
	}

	c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Подключение к узлу прошло успешно", slog.String("node", nodeName.String()))
	fmt.Printf("Подключение к узлу %s прошло успешно\n", nodeName)

	return nil
}

func (c *Cli) syncCmd(cmd *cobra.Command, args []string) error {
	c.logger.Debug("Выполнение команды sync")

	rootName, err := domain.NewRootName(args[0])
	if err != nil {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Не удалось выполнить sync", slog.String("error", err.Error()))
		fmt.Println("Не удалось выполнить sync")
		return err
	}

	cmdCtx := cmd.Context()
	interruptCtx, interruptCancel := signal.NotifyContext(cmdCtx, os.Interrupt)
	defer interruptCancel()

	shouldUseCache := *c.shouldUseCache
	appliedChanges, conflicts, baseSnapshotSaveError, userDecision, err := c.syncUseCase.ApplySyncChanges(interruptCtx, shouldUseCache, rootName)
	if err != nil {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Не удалось выполнить sync", slog.String("error", err.Error()))
		return err
	}

	var wasErr bool
	var i int
	go func() {
		for changeEvent := range appliedChanges {
			i++
			changeEventErr := changeEvent.Err
			if changeEventErr != nil {
				wasErr = true
				c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Не удалось применить изменение", slog.String("error", changeEventErr.Error()))
				fmt.Printf("Не удалось применить изменение: %s\n", changeEventErr)
				continue
			}
			c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Применено изменение", slog.String("change", c.changeToHumanReadable(i, changeEvent.Change)))
			fmt.Printf("Применено изменение:\n%s\n", c.changeToHumanReadable(0, changeEvent.Change))
		}
	}()

	for conflict := range conflicts {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Обнаружен конфликт", slog.Any("conflict", conflict))
		fmt.Println(c.conflictToHumanReadable(conflict))
		prompt := promptui.Select{
			Label: "Выберите решение для конфликта",
			Items: []string{
				c.decisionToHumanReadable(domain.LocalWin),
				c.decisionToHumanReadable(domain.RemoteWin),
				c.decisionToHumanReadable(domain.Skip),
			},
		}

		_, decision, err := prompt.Run()
		if err != nil {
			wasErr = true
			c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Ошибка при выборе решения конфликта", slog.String("error", err.Error()))
			return err
		}

		userDecision <- c.fromHumanReadableDecision(decision)
	}

	if err := <-baseSnapshotSaveError; err != nil {
		wasErr = true
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Ошибка при сохранении базового снимка", slog.String("error", err.Error()))
		return err
	}

	if interruptCtx.Err() != nil {
		c.logger.Debug("Синхронизация прервана пользователем")
		fmt.Println("Синхронизация прервана пользователем")
		return nil
	}

	if wasErr {
		c.logger.Debug("Синхронизация завершена с ошибками")
		fmt.Println("Синхронизация завершена с ошибками")
		return nil
	}

	c.logger.Debug("Синхронизация завершена")
	fmt.Println("Синхронизация завершена")

	return nil
}

func (c *Cli) decisionToHumanReadable(decision domain.Decision) string {
	switch decision {
	case domain.LocalWin:
		return decisionLocalWin
	case domain.RemoteWin:
		return decisionRemoteWin
	case domain.Skip:
		return decisionSkip
	default:
		return decisionUnknown
	}
}

func (c *Cli) fromHumanReadableDecision(decision string) domain.Decision {
	switch decision {
	case decisionLocalWin:
		return domain.LocalWin
	case decisionRemoteWin:
		return domain.RemoteWin
	case decisionSkip:
		return domain.Skip
	default:
		return -1
	}
}

func (c *Cli) dryRunCmd(cmd *cobra.Command, args []string) error {
	c.logger.Debug("Выполнение команды dry-run")
	rootName, err := domain.NewRootName(args[0])
	if err != nil {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Не удалось выполнить dry-run", slog.String("error", err.Error()))
		fmt.Println("Не удалось выполнить dry-run")
		return err
	}

	cmdCtx := cmd.Context()
	interruptCtx, interruptCancel := signal.NotifyContext(cmdCtx, os.Interrupt)
	defer interruptCancel()

	plan, err := c.scanUseCase.PlanSyncChanges(interruptCtx, rootName)
	if err != nil {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Не удалось получить изменения для синхронизации", slog.String("error", err.Error()))
		fmt.Print("Не удалось получить изменения для синхронизации\n")
		return err
	}

	if plan.IsEmpty() {
		c.logger.Debug("Нет изменений для синхронизации")
		fmt.Print("Нет изменений для синхронизации\n")
		return nil
	}

	changesHeader := c.getChangesHeader()

	if plan.IsAnyLocalChange() {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Найдены локальные изменения", slog.Int("count", plan.LocalLength()))
		fmt.Printf("Всего локальных изменений: %d\n%s\n", plan.LocalLength(), changesHeader)
		localChanges := plan.LocalChanges
		for i, change := range localChanges {
			fmt.Println(c.changeToHumanReadable(i+1, change.ToSyncChange()))
		}
	}

	if plan.IsAnyRemoteChange() {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Найдены удалённые изменения", slog.Int("count", plan.RemoteLength()))
		fmt.Printf("Всего удалённых изменений: %d\n%s\n", plan.RemoteLength(), changesHeader)
		remoteChanges := plan.RemoteChanges
		for i, change := range remoteChanges {
			fmt.Println(c.changeToHumanReadable(i+1, change.ToSyncChange()))
		}
	}

	if plan.IsAnyConflict() {
		c.logger.LogAttrs(c.loggerCtx, slog.LevelDebug, "Найдены конфликты", slog.Int("count", plan.ConflictLength()))
		fmt.Printf("Всего конфликтов: %d\n", plan.ConflictLength())
		conflicts := plan.Conflicts
		for i, conflict := range conflicts {
			fmt.Printf("%d.%s\n", i+1, c.conflictToHumanReadable(conflict))
		}
	}

	return nil
}

func (c *Cli) getChangesHeader() string {
	return fmt.Sprintf("%-4s %-12s %-40s %-40s\n%s",
		"#",
		"TYPE",
		"OLD PATH",
		"NEW PATH",
		strings.Repeat("-", 100),
	)
}

func (c *Cli) changeToHumanReadable(index int, s domain.SyncChange) string {
	rank := ""
	if index > 0 {
		rank = fmt.Sprintf("%d", index)
	}

	oldPath := c.pathOrDash(s.OldRelativePath)
	newPath := c.pathOrDash(s.NewRelativePath)

	switch s.ChangeType {
	case domain.CreateFile, domain.CreateDir:
		return fmt.Sprintf("%-4s %-12s %-40s %-40s",
			rank,
			changeTypeCreate,
			oldPath,
			newPath,
		)
	case domain.Delete:
		return fmt.Sprintf("%-4s %-12s %-40s %-40s",
			rank,
			changeTypeDelete,
			oldPath,
			newPath,
		)
	case domain.Rename:
		return fmt.Sprintf("%-4s %-12s %-40s %-40s",
			rank,
			changeTypeRename,
			oldPath,
			newPath,
		)
	case domain.Move:
		return fmt.Sprintf("%-4s %-12s %-40s %-40s",
			rank,
			changeTypeMove,
			oldPath,
			newPath,
		)
	case domain.Modify:
		return fmt.Sprintf("%-4s %-12s %-40s %-40s",
			rank,
			changeTypeModify,
			oldPath,
			newPath,
		)
	default:
		return fmt.Sprintf("%-4s %-12s %-40s %-40s",
			rank,
			changeTypeUnknown,
			oldPath,
			newPath,
		)
	}
}

func (c *Cli) conflictToHumanReadable(s domain.Conflict) string {
	var conflictType string

	switch s.Conflict {
	case domain.ConflictLocalDeletedRemoteModified:
		conflictType = conflictTypeLocalDeletedRemoteModified

	case domain.ConflictRemoteDeletedLocalModified:
		conflictType = conflictTypeRemoteDeletedLocalModified

	case domain.ConflictLocalMovedRemoteMoved:
		conflictType = conflictTypeLocalMovedRemoteMoved

	case domain.ConflictLocalRenamedRemoteRenamed:
		conflictType = conflictTypeLocalRenamedRemoteRenamed

	case domain.ConflictBothModifiedAtSameTime:
		conflictType = conflictTypeBothModifiedAtSameTime

	case domain.ConflictBothCreatedAtSamePathConflict:
		conflictType = conflictTypeBothCreatedAtSamePathConflict

	default:
		conflictType = conflictTypeUnknown
	}

	return fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n",
		strings.Repeat("=", 84),
		fmt.Sprintf("  Конфликт:                             %s", conflictType),
		fmt.Sprintf("  Локальный путь:                       %s", c.pathOrDash(s.LocalRelativePath)),
		fmt.Sprintf("  Удалённый путь:                       %s", c.pathOrDash(s.RemoteRelativePath)),
		fmt.Sprintf("  Базовый путь:                         %s", c.pathOrDash(s.BaseRelativePath)),
		fmt.Sprintf("  Локально изменён:                     %s", c.formatConflictTimestamp(s.LocalModifiedUnix)),
		fmt.Sprintf("  Удалённо изменён:                     %s", c.formatConflictTimestamp(s.RemoteModifiedUnix)),
		fmt.Sprintf("  Время изменения по базовому снимку:   %s", c.formatConflictTimestamp(s.BaseModifiedUnix)),
		strings.Repeat("=", 84),
	)
}

func (c *Cli) pathOrDash(path domain.Path) string {
	if path.IsEmpty() {
		return "-"
	}
	return path.String()
}

func (c *Cli) formatConflictTimestamp(timestamp uint64) string {
	if timestamp == 0 {
		return "-"
	}
	return time.Unix(int64(timestamp), 0).Format("2006-01-02 15:04:05")
}

func (c *Cli) suggestionFunc(comand *cobra.Command, annotationValue string, document *prompt.Document) []prompt.Suggest {
	typedPrefix := document.TextBeforeCursor()

	switch annotationValue {
	case connectCmdName:
		return c.nodeNameSuggestionFunc(typedPrefix)
	case dryRunCmdName, syncCmdName:
		return c.rootNameSuggestionFunc(typedPrefix)
	case initCmdName:
		return c.initNameSuggestionFunc(typedPrefix)
	case setAliasCmdName:
		return c.setAliasSuggestionFunc(typedPrefix)
	case removeAliasCmdName:
		return c.removeAliasSuggestionFunc(typedPrefix)
	case removeRootCmdName:
		return c.rootNameSuggestionFunc(typedPrefix)
	default:
		return nil
	}
}

func (c *Cli) setAliasSuggestionFunc(prefix string) []prompt.Suggest {
	suggestions := make([]prompt.Suggest, 0)

	parts := strings.Fields(prefix)

	if strings.HasSuffix(prefix, " ") {
		parts = append(parts, "")
	}

	length := len(parts)

	switch length {
	case 2:
		nodeNames := c.nodeNameCache
		for _, nodeName := range nodeNames {
			if strings.HasPrefix(nodeName.String(), parts[1]) {
				suggestions = append(suggestions, prompt.Suggest{
					Text: nodeName.String(),
				})
			}
		}
	case 3:
		return nil
	}

	return suggestions
}

func (c *Cli) removeAliasSuggestionFunc(prefix string) []prompt.Suggest {
	suggestions := make([]prompt.Suggest, 0)

	parts := strings.Fields(prefix)

	if strings.HasSuffix(prefix, " ") {
		parts = append(parts, "")
	}

	length := len(parts)

	if length < 2 {
		return nil
	}

	aliasByNodeName := c.aliasUseCase.GetAliases()

	for nodeName, alias := range aliasByNodeName {
		if strings.HasPrefix(alias, parts[1]) {
			suggestions = append(suggestions, prompt.Suggest{
				Text:        alias,
				Description: nodeName.String(),
			})
		}
	}

	return suggestions
}

func (c *Cli) nodeNameSuggestionFunc(prefix string) []prompt.Suggest {
	parts := strings.Fields(prefix)

	if strings.HasSuffix(prefix, " ") {
		parts = append(parts, "")
	}

	length := len(parts)

	if length < 2 {
		return nil
	}
	suggestions := make([]prompt.Suggest, 0)
	alreadyAdded := make(map[string]struct{})
	if len(c.nodeNameCache) != 0 {
		for _, nodeName := range c.nodeNameCache {
			nodeNameCandidate := nodeName.String()
			description := ""
			if alias, ok := c.aliasUseCase.GetAliasByNode(nodeName); ok {
				alreadyAdded[nodeNameCandidate] = struct{}{}
				nodeNameCandidate = alias
				description = nodeName.String()
			}

			if strings.HasPrefix(nodeNameCandidate, parts[1]) {
				suggestions = append(suggestions, prompt.Suggest{
					Text:        nodeNameCandidate,
					Description: description,
				})
			}
		}
	}

	for nodeName, alias := range c.aliasUseCase.GetAliases() {
		if _, ok := alreadyAdded[nodeName.String()]; ok {
			continue
		}
		nodeNameCandidate := nodeName.String()
		descripton := ""
		if alias != "" {
			nodeNameCandidate = alias
			descripton = nodeName.String()
		}
		if strings.HasPrefix(nodeNameCandidate, parts[1]) {
			suggestions = append(suggestions, prompt.Suggest{
				Text:        nodeNameCandidate,
				Description: descripton,
			})
		}

	}

	return suggestions
}

func (c *Cli) initNameSuggestionFunc(prefix string) []prompt.Suggest {
	suggestions := make([]prompt.Suggest, 0)

	parts := strings.Fields(prefix)

	if strings.HasSuffix(prefix, " ") {
		parts = append(parts, "")
	}

	length := len(parts)

	if length != 2 {
		return nil
	}

	rootNames, err := c.rootUseCase.GetRoots()
	if err != nil {
		return suggestions
	}

	for rootName, path := range rootNames {
		if strings.HasPrefix(rootName.String(), parts[1]) {
			suggestions = append(suggestions, prompt.Suggest{
				Text:        rootName.String(),
				Description: path.String(),
			})
		}
	}

	return suggestions
}

func (c *Cli) rootNameSuggestionFunc(prefix string) []prompt.Suggest {
	suggestions := make([]prompt.Suggest, 0)

	parts := strings.Fields(prefix)

	if strings.HasSuffix(prefix, " ") {
		parts = append(parts, "")
	}

	length := len(parts)

	if length != 2 {
		return nil
	}

	rootNames, err := c.rootUseCase.GetRoots()
	if err != nil {
		return suggestions
	}

	for rootName, path := range rootNames {
		if strings.HasPrefix(rootName.String(), parts[1]) {
			suggestions = append(suggestions, prompt.Suggest{
				Text:        rootName.String(),
				Description: path.String(),
			})
		}
	}

	if len(suggestions) == 0 {
		for rootName, path := range rootNames {
			suggestions = append(suggestions, prompt.Suggest{
				Text:        rootName.String(),
				Description: path.String(),
			})
		}
	}

	return suggestions
}
