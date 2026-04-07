package cli

import (
	"context"
	"insync/internal/interfaces"
	"log/slog"

	"github.com/spf13/cobra"
)

type Cli struct {
	logger    *slog.Logger
	loggerCtx context.Context

	syncUseCase interfaces.ISyncUseCase
}

type CliOptions struct {
	Logger *slog.Logger
}

func NewCli(opts CliOptions) interfaces.ICli {
	if opts.Logger == nil {
		panic("Не все обязательные параметры были переданы при инициализации Cli")
	}
	loggerCtx := context.Background()
	return &Cli{
		logger:    opts.Logger,
		loggerCtx: loggerCtx,
	}
}

func (c *Cli) Start() error {
	rootCmd := &cobra.Command{
		Use:   "insync",
		Short: "InSync - инструмент для синхронизации файлов между различными хранилищами",
	}

	return rootCmd.Execute()
}
