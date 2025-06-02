package createuser

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/auth"
	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/database"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"golang.org/x/term"
)

var CreateUserCmd = &cobra.Command{
	Use:   "createuser",
	Short: "Create a new CastKeeper user",
	Long:  "Utility script for creating new users in the database for the given CastKeeper configuration.",
	Run:   run,
}

var (
	cfgFile  string
	username string
	password string
)

func init() {
	CreateUserCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (otherwise uses default locations)")
	CreateUserCmd.Flags().StringVar(&username, "username", "", "username for the user to create (otherwise entered interactively)")
	CreateUserCmd.Flags().StringVar(&password, "password", "", "password for the user to create (otherwise entered interactively)")
}

func run(cmd *cobra.Command, args []string) {
	cfg, logger, err := config.LoadConfig(cfgFile)
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
	argUsername := username
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
	argPassword := password
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
