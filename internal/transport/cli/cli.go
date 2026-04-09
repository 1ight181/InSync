package cli

import (
	"context"
	"fmt"
	"insync/internal/domain"
	"insync/internal/interfaces"
	"log/slog"

	"github.com/spf13/cobra"
)

type Cli struct {
	logger    *slog.Logger
	loggerCtx context.Context

	syncUseCase    interfaces.ISyncUseCase
	connectUseCase interfaces.IConnectUseCase
}

type CliOptions struct {
	Logger *slog.Logger

	SyncUseCase    interfaces.ISyncUseCase
	ConnectUseCase interfaces.IConnectUseCase
}

func NewCli(opts CliOptions) interfaces.ICli {
	if opts.Logger == nil {
		panic("Не все обязательные параметры были переданы при инициализации Cli")
	}
	loggerCtx := context.Background()
	return &Cli{
		logger:    opts.Logger,
		loggerCtx: loggerCtx,

		syncUseCase:    opts.SyncUseCase,
		connectUseCase: opts.ConnectUseCase,
	}
}

func (c *Cli) Start() error {
	rootCmd := c.createRootCmd()
	nodesCmd := c.createNodesCmd()
	connectCmd := c.createConnectCmd()
	dryRunCmd := c.createConnectCmd()
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
		Run:   c.nodesCommand,
	}
}

func (c *Cli) createConnectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "connect node-name",
		Short: "Подключиться к узлу",
		Long: `Подключиться к узлу. 
		node-name - имя узла, к которому будет осуществляться подключение`,
		RunE: c.connectCommand,
		Args: cobra.ExactArgs(1),
	}
}

func (c *Cli) createDryRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dry-run root-name",
		Short: "Показать, какие файлы/каталоги будут синхронизированы, без фактического выполнения синхронизации",
		Run:   c.dryRunCommand,
		Args:  cobra.ExactArgs(1),
	}
}

func (c *Cli) createSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync root-name",
		Short: "Выполнить синхронизацию",
		RunE:  c.syncCommand,
		Args:  cobra.ExactArgs(1),
	}
}

func (c *Cli) nodesCommand(cmd *cobra.Command, args []string) {
	c.logger.Debug("Выполнение команды nodes")
	nodes, err := c.connectUseCase.GetAllNodes()
	if err != nil {
		c.logger.Error("Не удалось получить узлы", "error", err)
		return
	}

	if len(nodes) == 0 {
		c.logger.Info("Нет доступных узлов")
		return
	}

	c.logger.Info("Доступные узлы:")

	for i, node := range nodes {
		c.logger.Info(fmt.Sprintf("%d. %s", i+1, node.Name))
	}
}

func (c *Cli) connectCommand(cmd *cobra.Command, args []string) error {
	c.logger.Debug("Выполнение команды connect")

	nodeName := args[0]
	c.logger.Info(fmt.Sprintf("Подключение к узлу %s", nodeName))

	if err := c.connectUseCase.ConnectToNode(nodeName); err != nil {
		return err
	}

	c.logger.Info(fmt.Sprintf("Подключение к узлу %s прошло успешно", nodeName))

	return nil
}

func (c *Cli) syncCommand(cmd *cobra.Command, args []string) error {
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
		c.logger.Info(fmt.Sprintf("Применено изменение %s:\n %s", c.changeToHumanReadable(changeEvent.Change)))
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

func (c *Cli) dryRunCommand(cmd *cobra.Command, args []string) {
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
