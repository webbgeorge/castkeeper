package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/database"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"gorm.io/gorm"
)

var (
	cfgFile string
	verbose bool
)

func InitGlobalFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (otherwise uses default locations)")
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
}

func ConfigureCLI() (context.Context, config.Config, *gorm.DB, error) {
	logLevel := slog.LevelWarn
	_ = os.Setenv("CASTKEEPER_LOGLEVEL", "warn")
	if verbose {
		logLevel = slog.LevelDebug
		_ = os.Setenv("CASTKEEPER_LOGLEVEL", "debug")
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	ctx := framework.ContextWithLogger(context.Background(), logger)

	cfg, _, err := config.LoadConfig(cfgFile)
	if err != nil {
		return nil, config.Config{}, nil, fmt.Errorf("failed to read config: %w", err)
	}

	db, err := database.ConfigureDatabase(cfg, logger, false)
	if err != nil {
		return nil, config.Config{}, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return ctx, cfg, db, nil
}
