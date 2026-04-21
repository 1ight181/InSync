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

	prompt "github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	cobraprompt "github.com/stromland/cobra-prompt"

	promptui "github.com/manifoldco/promptui"
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

	shouldUseCache *bool
	nodeNameCache  []domain.NodeName
}

type CliOptions struct {
	SyncUseCase    ISyncUseCase
	ScanUseCase    IScanUseCase
	NodeUseCase    INodeUseCase
	ConnectUseCase IConnectUseCase
	RootUseCase    IRootUseCase

	Logger *slog.Logger
}

var (
	ErrInvalidCliOptions = errors.New("Все поля CliOptions должны быть заполнены")
)

func NewCli(opts CliOptions) (*Cli, error) {
	if opts.SyncUseCase == nil ||
		opts.ScanUseCase == nil ||
		opts.NodeUseCase == nil ||
		opts.ConnectUseCase == nil ||
		opts.RootUseCase == nil ||
		opts.Logger == nil {
		return nil, ErrInvalidCliOptions
	}
	loggerCtx := context.Background()
	return &Cli{
		syncUseCase:    opts.SyncUseCase,
		scanUseCase:    opts.ScanUseCase,
		nodeUseCase:    opts.NodeUseCase,
		connectUseCase: opts.ConnectUseCase,
		rootUseCase:    opts.RootUseCase,

		logger:        opts.Logger,
		loggerCtx:     loggerCtx,
		nodeNameCache: make([]domain.NodeName, 0),
	}, nil
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

	cmdCtx := cmd.Context()
	interruptCtx, interruptCancel := signal.NotifyContext(cmdCtx, os.Interrupt)
	defer interruptCancel()

	nodeNamesChan, err := c.nodeUseCase.ShowNodeNames(interruptCtx)
	if err != nil {
		c.logger.Error("Не удалось получить узлы", "error", err)
		return
	}

	c.logger.Info("Доступные узлы (динамический список, нажмите Ctrl+C для завершения):")

	i := 1
	select {
	case nodeName := <-nodeNamesChan:
		c.logger.Info(fmt.Sprintf("%d. %s", i, nodeName))
		i++
		c.nodeNameCache = append(c.nodeNameCache, nodeName)
	case <-interruptCtx.Done():
		return
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
	appliedChanges, conflicts, baseSnapshotSaveError, userDecision, err := c.syncUseCase.ApplySyncChanges(interruptCtx, shouldUseCache, rootName)
	if err != nil {
		return err
	}

	go func() {
		for changeEvent := range appliedChanges {

			changeEventErr := changeEvent.Err
			if changeEventErr != nil {
				c.logger.Error(fmt.Sprintf("Не удалось применить изменение: %s", changeEventErr))
				continue
			}
			c.logger.Info(fmt.Sprintf("Применено изменение: %s\n", c.changeToHumanReadable(changeEvent.Change)))
		}
	}()

	for conflict := range conflicts {
		conflictLabel := c.conflictToHumanReadable(conflict)
		prompt := promptui.Select{
			Label: conflictLabel,
			Items: []string{
				c.decisionToHumanReadable(domain.LocalWin),
				c.decisionToHumanReadable(domain.RemoteWin),
				c.decisionToHumanReadable(domain.Skip),
			},
		}

		_, decision, err := prompt.Run()
		if err != nil {
			return err
		}

		userDecision <- c.fromHumanReadableDecision(decision)
	}

	select {
	case err := <-baseSnapshotSaveError:
		return err
	default:
	}

	if interruptCtx.Err() != nil {
		c.logger.Error("Синхронизация прервана пользователем")
		return nil
	}

	c.logger.Info("Синхронизация завершена")

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
		conflictType = conflictTypeBothModifiedAtSameTime

	case domain.ConflictBothCreatedAtSamePathConflict:
		conflictType = conflictTypeBothCreatedAtSamePathConflict

	default:
		conflictType = conflictTypeUnknown
	}

	return fmt.Sprintf(
		`Тип конфликта: %s
		Файл на локальном узле: %s
		Файл на удалённом узле: %s
		Файл на базовом снимке: %s
		Последний раз модифицирован на локальном узле: %d
		Последний раз модифицирован на удалённом узле: %d
		Последний раз модифицирован по базовому снимку: %d`,
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
	suggestions := make([]prompt.Suggest, 0)
	for _, nodeName := range c.nodeNameCache {
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
