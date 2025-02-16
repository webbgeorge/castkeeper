package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	slogGorm "github.com/orandin/slog-gorm"
	"github.com/webbgeorge/castkeeper/pkg/auth"
	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	configFile := "" // optional specific config file (otherwise uses default locations)
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	cfg, logger, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	ctx := framework.ContextWithLogger(context.Background(), logger)

	db, err := configureDatabase(cfg, logger)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("failed to read username: %v", err)
	}

	fmt.Print("Enter password: ")
	password, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("failed to read password: %v", err)
	}

	err = auth.CreateUser(ctx, db, strings.TrimSpace(username), strings.TrimSpace(password))
	if err != nil {
		log.Fatalf("failed to create user: %v", err)
	}

	log.Printf("successfully created user '%s'", username)
}

// TODO this is a copy, needs to be moved somewhere central
func configureDatabase(cfg config.Config, logger *slog.Logger) (*gorm.DB, error) {
	gormLogger := slogGorm.New(
		slogGorm.WithHandler(logger.Handler()),
	)

	db, err := gorm.Open(
		dbDialector(cfg),
		&gorm.Config{
			TranslateError: true,
			Logger:         gormLogger,
		},
	)
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&podcasts.Podcast{}); err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&podcasts.Episode{}); err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&auth.User{}); err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&auth.Session{}); err != nil {
		return nil, err
	}

	return db, nil
}

func dbDialector(cfg config.Config) gorm.Dialector {
	switch cfg.Database.Driver {
	case config.DatabaseDriverPostgres:
		return postgres.New(postgres.Config{
			DSN:                  cfg.Database.DSN,
			PreferSimpleProtocol: true,
		})
	case config.DatabaseDriverSqlite:
		return sqlite.Open(cfg.Database.DSN)
	default:
		return nil
	}
}
