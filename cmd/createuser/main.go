package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/webbgeorge/castkeeper/pkg/auth"
	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/database"
	"github.com/webbgeorge/castkeeper/pkg/framework"
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

	db, err := database.ConfigureDatabase(cfg, logger)
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
