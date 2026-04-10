package cli

import (
	"context"
	"fmt"
	"insync/internal/domain"
	"log/slog"

	"github.com/spf13/cobra"
)

type Cli struct {
	syncUseCase    ISyncUseCase
	connectUseCase IConnectUseCase

	rootUseCase *IRootUseCase

	logger    *slog.Logger
	loggerCtx context.Context

	ctx context.Context
}

type CliOptions struct {
	SyncUseCase    ISyncUseCase
	ConnectUseCase IConnectUseCase

	Logger *slog.Logger

	Ctx context.Context
}

func NewCli(opts CliOptions) *Cli {
	if opts.SyncUseCase == nil ||
		opts.ConnectUseCase == nil ||
		opts.Logger == nil ||
		opts.Ctx == nil {
		panic("Не все обязательные параметры были переданы при инициализации Cli")
	}
	loggerCtx := context.Background()
	return &Cli{
		syncUseCase:    opts.SyncUseCase,
		connectUseCase: opts.ConnectUseCase,

		logger:    opts.Logger,
		loggerCtx: loggerCtx,

		ctx: opts.Ctx,
	}
}

func (c *Cli) Start() error {
	rootCmd := c.createRootCmd()
	nodesCmd := c.createNodesCmd()
	connectCmd := c.createConnectCmd()
	dryRunCmd := c.createDryRunCmd()
	syncCmd := c.createSyncCmd()

	rootCmd.AddCommand(dryRunCmd)
	rootCmd.AddCommand(nodesCmd)
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(syncCmd)

	return rootCmd.Execute()
}

func (c *Cli) createRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "insync",
		Short: "InSync - инструмент для синхронизации файлов между различными хранилищами",
	}
}

func (c *Cli) createNodesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "nodes",
		Short: "Отобразить доступные для синхронизации узлы",
		Args:  cobra.NoArgs,
		Run:   c.nodesCmd,
	}
}

func (c *Cli) createConnectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "connect node-name",
		Short: "Подключиться к узлу",
		Long: `Подключиться к узлу. 
		node-name - имя узла, к которому будет осуществляться подключение`,
		RunE: c.connectCmd,
		Args: cobra.ExactArgs(1),
	}
}

func (c *Cli) createDryRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dry-run root-name",
		Short: "Показать, какие файлы/каталоги будут синхронизированы, без фактического выполнения синхронизации",
		Run:   c.dryRunCmd,
		Args:  cobra.ExactArgs(1),
	}
}

func (c *Cli) createSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync root-name",
		Short: "Выполнить синхронизацию",
		RunE:  c.syncCmd,
		Args:  cobra.ExactArgs(1),
	}
}

func (c *Cli) nodesCmd(cmd *cobra.Command, args []string) {
	c.logger.Debug("Выполнение команды nodes")
	nodeNamesChan, err := c.connectUseCase.ShowNodes()
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

func (c *Cli) connectCmd(cmd *cobra.Command, args []string) error {
	c.logger.Debug("Выполнение команды connect")

	nodeName := args[0]
	c.logger.Info(fmt.Sprintf("Подключение к узлу %s", nodeName))

	if err := c.connectUseCase.ConnectToNode(nodeName); err != nil {
		return err
	}

	c.logger.Info(fmt.Sprintf("Подключение к узлу %s прошло успешно", nodeName))

	return nil
}

func (c *Cli) syncCmd(cmd *cobra.Command, args []string) error {
	c.logger.Debug("Выполнение команды sync")

	changes := c.syncUseCase.GetSyncChanges()
	changesLen := len(changes)
	if changesLen == 0 {
		c.logger.Info("Нет изменений для синхронизации")
	}

	changeEventChan, err := c.syncUseCase.ApplySyncChanges(changes)
	if err != nil {
		return err
	}

	changeCount := 0

	for changeEvent := range changeEventChan {
		changeCount++
		changeEventErr := changeEvent.Err
		if changeEventErr != nil {
			return changeEventErr
		}
		c.logger.Info(fmt.Sprintf("Применено изменение: %s\n", c.changeToHumanReadable(changeEvent.Change)))
		if changeCount > changesLen {
			c.logger.LogAttrs(
				c.loggerCtx,
				slog.LevelWarn,
				"Необычное поведение: применено изменений больше чем было завялено",
				slog.Int("declared", changesLen),
				slog.Int("applied", changeCount),
			)
		}
	}

	if changeCount < changesLen {
		c.logger.LogAttrs(
			c.loggerCtx,
			slog.LevelWarn,
			"Необычное поведение: применено изменений меньше чем было завялено, несмотря на отсутствие ошибок",
			slog.Int("declared", changesLen),
			slog.Int("applied", changeCount),
		)

	}

	c.logger.Info("Синхронизация завершена")

	return nil
}

func (c *Cli) dryRunCmd(cmd *cobra.Command, args []string) {
	c.logger.Debug("Выполнение команды dry-run")
	changes := c.syncUseCase.GetSyncChanges()
	changesLen := len(changes)
	if changesLen == 0 {
		c.logger.Info("Нет изменений для синхронизации")
	}
	changesHeader := c.getChangesHeader()
	c.logger.Info(fmt.Sprintf("Всего изменений: %d\n%s", changesLen, changesHeader))
	for i, change := range changes {
		c.logger.Info(fmt.Sprintf("%d. %s", i+1, c.changeToHumanReadable(change)))
	}
}

func (c *Cli) getChangesHeader() string {
	return fmt.Sprintf("|%s|%s|%s|%s|\n", "CHANGE TYPE", "ROOT NAME", "OLD RELATIVE PATH", "NEW RELATIVE PATH")
}

func (c *Cli) changeToHumanReadable(s domain.SyncChange) string {
	switch s.ChangeType {
	case domain.Create:
		return fmt.Sprintf("|CREATE|%s|%s|\n", s.RootName, s.NewRelativePath)
	case domain.Delete:
		return fmt.Sprintf("|DELETE|%s|%s|\n", s.RootName, s.OldRelativePath)
	case domain.Rename:
		return fmt.Sprintf("|RENAME|%s|%s|%s|\n", s.RootName, s.OldRelativePath, s.NewRelativePath)
	case domain.Move:
		return fmt.Sprintf("|MOVE|%s|%s|%s|\n", s.RootName, s.OldRelativePath, s.NewRelativePath)
	default:
		return "UNKNOWN CHANGE TYPE"
	}

}
