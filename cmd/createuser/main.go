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
	"golang.org/x/term"
)

// Utility script for creating new users in the database for the given CastKeeper configuration.
//
// Usage:
//
//	go run cmd/createuser [configPath] [username] [password]
//
// accepts interactive inputs when username and password are not given as args
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

	username, err := readUsername()
	if err != nil {
		log.Fatalf("failed to read username: %v", err)
	}

	password, err := readPassword()
	if err != nil {
		log.Fatalf("failed to read password: %v", err)
	}

	err = auth.CreateUser(ctx, db, username, password)
	if err != nil {
		log.Fatalf("failed to create user: %v", err)
	}

	log.Printf("successfully created user '%s'", username)
}

func readUsername() (string, error) {
	argUsername := readArg(2)
	if argUsername != "" {
		return argUsername, nil
	}

	fmt.Print("Enter username: ")
	username, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(username), nil
}

func readPassword() (string, error) {
	argPassword := readArg(3)
	if argPassword != "" {
		return argPassword, nil
	}

	fmt.Print("Enter password: ")
	pwBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	return string(pwBytes), nil
}

func readArg(i int) string {
	if len(os.Args) > i {
		return strings.TrimSpace(os.Args[i])
	}
	return ""
}
