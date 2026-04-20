package cli

import (
	"context"
	"fmt"
	"insync/internal/domain"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	cobraprompt "github.com/stromland/cobra-prompt"
)

const (
	rootCmdName        = "root"
	nodesCmdName       = "nodes"
	currentNodeCmdName = "current-node"
	addRootCmdName     = "add-root"
	removeRootCmdName  = "remove-root"
	getRootsCmdName    = "get-roots"
	connectCmdName     = "connect"
	dryRunCmdName      = "dry-run"
	syncCmdName        = "sync"
)

const (
	conflictTypeLocalDeletedRemoteModified     = "LOCAL_DELETED_REMOTE_MODIFIED"
	conflictTypeRemoteDeletedLocalModified     = "REMOTE_DELETED_LOCAL_MODIFIED"
	conflictTypeLocalMovedRemoteMoved          = "LOCAL_MOVED_REMOTE_MOVED"
	conflictTypeLocalRenamedRemoteRenamed      = "LOCAL_RENAMED_REMOTE_RENAMED"
	conflictTypeBothModifiedConflictAtSameTime = "BOTH_MODIFIED_CONFLICT_AT_SAME_TIME"
	conflictTypeBothCreatedAtSamePathConflict  = "BOTH_CREATED_AT_SAME_PATH_CONFLICT"
	conflictTypeUnknown                        = "UNKNOWN"
)

const (
	shouldUseCacheFlagName = "should_use_cache"
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

	logger    *slog.Logger
	loggerCtx context.Context

	ctx context.Context

	shouldUseCache *bool
}

type CliOptions struct {
	SyncUseCase    ISyncUseCase
	ScanUseCase    IScanUseCase
	NodeUseCase    INodeUseCase
	ConnectUseCase IConnectUseCase
	RootUseCase    IRootUseCase

	Logger *slog.Logger

	Ctx context.Context
}

func NewCli(opts CliOptions) *Cli {
	if opts.SyncUseCase == nil ||
		opts.ScanUseCase == nil ||
		opts.NodeUseCase == nil ||
		opts.ConnectUseCase == nil ||
		opts.RootUseCase == nil ||
		opts.Logger == nil ||
		opts.Ctx == nil {
		panic("Не все обязательные параметры были переданы при инициализации Cli")
	}
	loggerCtx := context.Background()
	return &Cli{
		syncUseCase:    opts.SyncUseCase,
		scanUseCase:    opts.ScanUseCase,
		nodeUseCase:    opts.NodeUseCase,
		connectUseCase: opts.ConnectUseCase,
		rootUseCase:    opts.RootUseCase,

		logger:    opts.Logger,
		loggerCtx: loggerCtx,

		ctx: opts.Ctx,
	}
}

func (c *Cli) Start() {
	rootCmd := c.createRootCmd()

	nodesCmd := c.createNodesCmd()
	connectCmd := c.createConnectCmd()
	currentNodeCmd := c.createCurrentNodeCmd()

	addRootCmd := c.createAddRootCmd()
	removeRootCmd := c.createRemoveRootCmd()
	getRootsCmd := c.createGetRootsCmd()

	dryRunCmd := c.createDryRunCmd()
	syncCmd := c.createSyncCmd()

	rootCmd.AddCommand(nodesCmd)
	rootCmd.AddCommand(currentNodeCmd)

	rootCmd.AddCommand(addRootCmd)
	rootCmd.AddCommand(removeRootCmd)
	rootCmd.AddCommand(getRootsCmd)

	rootCmd.AddCommand(connectCmd)

	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(dryRunCmd)

	cobraPrompt := cobraprompt.CobraPrompt{
		RootCmd:                 rootCmd,
		ShowHelpCommandAndFlags: true,
		GoPromptOptions: []prompt.Option{
			prompt.OptionTitle("InSync 0.1"),
			prompt.OptionMaxSuggestion(5),
			prompt.OptionPrefix("insync> "),
			prompt.OptionInputTextColor(prompt.DarkGreen),
			prompt.OptionSelectedSuggestionTextColor(prompt.Green),
			prompt.OptionSuggestionTextColor(prompt.DarkGreen),
		},

		DynamicSuggestionsFunc: c.suggestionFunc,
	}

	cobraPrompt.Run()
}

func (c *Cli) createRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   rootCmdName,
		Short: "InSync - инструмент для синхронизации файлов между различными хранилищами",
	}
}

func (c *Cli) createNodesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   nodesCmdName,
		Short: "Отобразить доступные для синхронизации узлы",
		Args:  cobra.NoArgs,
		Run:   c.nodesCmd,
	}
}

func (c *Cli) createCurrentNodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   currentNodeCmdName,
		Short: "Отобразить текущий узел",
		Args:  cobra.NoArgs,
		Run:   c.currentNodeCmd,
	}
}

func (c *Cli) createAddRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s root-name root-path", addRootCmdName),
		Short: "Добавить корневой каталог",
		Long: `Добавить корневой каталог. 
		root-name - имя корневого каталога
		root-path - путь к корневому каталогу`,
		RunE: c.addRootCmd,
	}
}

func (c *Cli) createRemoveRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s root-name", removeRootCmdName),
		Short: "Удалить корневой каталог",
		Long: `Удалить корневой каталог. 
		root-name - имя корневого каталога`,
		RunE: c.removeRootCmd,
	}
}

func (c *Cli) createGetRootsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s", getRootsCmdName),
		Short: "Получить список корневых каталогов",
		Args:  cobra.NoArgs,
		Run:   c.getRootsCmd,
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
		RunE:  c.dryRunCmd,
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			cobraprompt.DynamicSuggestionsAnnotation: dryRunCmdName,
		},
	}
}

func (c *Cli) createSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s root-name", syncCmdName),
		Short: "Выполнить синхронизацию",
		RunE:  c.syncCmd,
		Args:  cobra.ExactArgs(1),
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

func (c *Cli) nodesCmd(cmd *cobra.Command, args []string) {
	c.logger.Debug("Выполнение команды nodes")
	nodeNamesChan, err := c.nodeUseCase.ShowNodeNames()
	if err != nil {
		c.logger.Error("Не удалось получить узлы", "error", err)
		return
	}

	c.logger.Info("Доступные узлы:")

	i := 1
	for nodeName := range nodeNamesChan {
		c.logger.Info(fmt.Sprintf("%d. %s", i, nodeName))
		i++
	}
}

func (c *Cli) currentNodeCmd(cmd *cobra.Command, args []string) {
	c.logger.Debug("Выполнение команды current-node")

	currentNode := c.connectUseCase.CurrentNodeName()
	c.logger.Info(fmt.Sprintf("Текущий подключенный узел: %s", currentNode))
}

func (c *Cli) addRootCmd(cmd *cobra.Command, args []string) error {
	c.logger.Debug("Выполнение команды add-root")

	rootName, err := domain.NewRootName(args[0])
	if err != nil {
		c.logger.Error("Не удалось создать корневой каталог")
		return err
	}

	rootRelativePath, err := domain.NewPath(args[1])
	if err != nil {
		c.logger.Error("Не удалось создать корневой каталог")
		return err
	}

	c.rootUseCase.AddRoot(rootName, rootRelativePath)

	c.logger.Info(fmt.Sprintf("Корневой каталог %s добавлен", rootName))
	return nil
}

func (c *Cli) removeRootCmd(cmd *cobra.Command, args []string) error {
	c.logger.Debug("Выполнение команды remove-root")

	rootName, err := domain.NewRootName(args[0])
	if err != nil {
		c.logger.Error("Не удалось создать удалить корневой каталог")
		return err
	}

	c.rootUseCase.RemoveRoot(rootName)

	c.logger.Info(fmt.Sprintf("Корневой каталог %s удален", rootName))

	return nil
}

func (c *Cli) getRootsCmd(cmd *cobra.Command, args []string) {
	c.logger.Debug("Выполнение команды get-roots")

	roots := c.rootUseCase.GetRoots()

	c.logger.Info("Список корневых каталогов:")

	for i, root := range roots {
		c.logger.Info(fmt.Sprintf("%d. %s", i+1, root))
	}
}

func (c *Cli) connectCmd(cmd *cobra.Command, args []string) error {
	c.logger.Debug("Выполнение команды connect")

	nodeName, err := domain.NewNodeName(args[0])
	if err != nil {
		c.logger.Error("Не удалось подключиться к узлу")
		return err
	}
	c.logger.Info(fmt.Sprintf("Подключение к узлу %s", nodeName))

	if err := c.connectUseCase.ConnectToNode(nodeName); err != nil {
		return err
	}

	c.logger.Info(fmt.Sprintf("Подключение к узлу %s прошло успешно", nodeName))

	return nil
}

func (c *Cli) syncCmd(cmd *cobra.Command, args []string) error {
	c.logger.Debug("Выполнение команды sync")

	rootName, err := domain.NewRootName(args[0])
	if err != nil {
		c.logger.Error("Не удалось выполнить sync")
		return err
	}

	cmdCtx := cmd.Context()
	interruptCtx, interruptCancel := signal.NotifyContext(cmdCtx, os.Interrupt)
	defer interruptCancel()

	shouldUseCache := *c.shouldUseCache
	changeEventChan, err := c.syncUseCase.ApplySyncChanges(interruptCtx, shouldUseCache, rootName)
	if err != nil {
		return err
	}

	for changeEvent := range changeEventChan {

		changeEventErr := changeEvent.Err
		if changeEventErr != nil {
			return changeEventErr
		}
		c.logger.Info(fmt.Sprintf("Применено изменение: %s\n", c.changeToHumanReadable(changeEvent.Change)))
	}

	c.logger.Info("Синхронизация завершена")

	return nil
}

func (c *Cli) dryRunCmd(cmd *cobra.Command, args []string) error {
	c.logger.Debug("Выполнение команды dry-run")
	rootName, err := domain.NewRootName(args[0])
	if err != nil {
		c.logger.Error("Не удалось выполнить dry-run")
		return err
	}

	cmdCtx := cmd.Context()
	interruptCtx, interruptCancel := signal.NotifyContext(cmdCtx, os.Interrupt)
	defer interruptCancel()

	plan, err := c.scanUseCase.PlanSyncChanges(interruptCtx, rootName)
	if err != nil {
		c.logger.Error("Не удалось получить изменения для синхронизации")
		return err
	}

	if plan.IsEmpty() {
		c.logger.Info("Нет изменений для синхронизации")
	}

	changesHeader := c.getChangesHeader()

	if plan.IsAnyLocalChange() {
		c.logger.Info(fmt.Sprintf("Всего локальных изменений: %d\n%s", plan.LocalLength(), changesHeader))
		localChanges := plan.LocalChanges
		for i, change := range localChanges {
			c.logger.Info(fmt.Sprintf("%d. %s", i+1, c.changeToHumanReadable(change.ToSyncChange())))
		}
	}

	if plan.IsAnyRemoteChange() {
		c.logger.Info(fmt.Sprintf("Всего удалённых изменений: %d\n%s", plan.RemoteLength(), changesHeader))
		remoteChanges := plan.RemoteChanges
		for i, change := range remoteChanges {
			c.logger.Info(fmt.Sprintf("%d. %s", i+1, c.changeToHumanReadable(change.ToSyncChange())))
		}
	}

	if plan.IsAnyConflict() {
		c.logger.Info(fmt.Sprintf("Всего конфликтов: %d\n%s", plan.ConflictLength(), changesHeader))
		conflicts := plan.Conflicts
		for i, conflict := range conflicts {
			c.logger.Info(fmt.Sprintf("%d. %s", i+1, c.conflictToHumanReadable(conflict)))
		}
	}

	return nil
}

func (c *Cli) getChangesHeader() string {
	return fmt.Sprintf("|%s|%s|%s|\n", "CHANGE TYPE", "OLD RELATIVE PATH", "NEW RELATIVE PATH")
}

func (c *Cli) changeToHumanReadable(s domain.SyncChange) string {
	switch s.ChangeType {
	case domain.Create:
		return fmt.Sprintf("|%s|%s|\n", changeTypeCreate, s.NewRelativePath)
	case domain.Delete:
		return fmt.Sprintf("|%s|%s|\n", changeTypeDelete, s.OldRelativePath)
	case domain.Rename:
		return fmt.Sprintf("|%s|%s|%s|\n", changeTypeRename, s.OldRelativePath, s.NewRelativePath)
	case domain.Move:
		return fmt.Sprintf("|%s|%s|%s|\n", changeTypeMove, s.OldRelativePath, s.NewRelativePath)
	case domain.Modify:
		return fmt.Sprintf("|%s|%s|%s|\n", changeTypeModify, s.OldRelativePath, s.NewRelativePath)
	default:
		return changeTypeUnknown
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
		conflictType = conflictTypeBothModifiedConflictAtSameTime

	case domain.ConflictBothCreatedAtSamePathConflict:
		conflictType = conflictTypeBothCreatedAtSamePathConflict

	default:
		conflictType = conflictTypeUnknown
	}

	return fmt.Sprintf(
		"Тип конфликта: %s\nLocalRelativePath : %s\nRemoteRelativePath : %s\nBaseRelativePath : %s\nLocalModifiedUnix : %d\nRemoteModifiedUnix : %d\nBaseModifiedUnix : %d\n",
		conflictType,
		s.LocalRelativePath,
		s.RemoteRelativePath,
		s.BaseRelativePath,
		s.LocalModifiedUnix,
		s.RemoteModifiedUnix,
		s.BaseModifiedUnix,
	)
}

func (c *Cli) suggestionFunc(annotationValue string, document *prompt.Document) []prompt.Suggest {
	typedPrefix := document.TextBeforeCursor()

	switch annotationValue {
	case nodesCmdName:
		c.nodeNameSuggestionFunc(typedPrefix)
	case dryRunCmdName, syncCmdName:
		c.rootNameSuggestionFunc(typedPrefix)
	default:
		return nil
	}

	return nil
}

func (c *Cli) nodeNameSuggestionFunc(prefix string) []prompt.Suggest {
	nodeNamesChan, err := c.nodeUseCase.ShowNodeNames()
	if err != nil {
		return nil
	}

	suggestions := make([]prompt.Suggest, 0)
	for nodeName := range nodeNamesChan {
		if strings.HasPrefix(nodeName.String(), prefix) {
			suggestions = append(suggestions, prompt.Suggest{
				Text: nodeName.String(),
			})
		}
	}

	return suggestions
}

func (c *Cli) rootNameSuggestionFunc(prefix string) []prompt.Suggest {
	rootNames := c.rootUseCase.GetRoots()
	suggestions := make([]prompt.Suggest, 0)
	for _, rootName := range rootNames {
		if strings.HasPrefix(rootName.String(), prefix) {
			suggestions = append(suggestions, prompt.Suggest{
				Text: rootName.String(),
			})
		}
	}

	return suggestions
}
