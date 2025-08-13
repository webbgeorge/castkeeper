package createuser

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/auth/users"
	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/database"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"golang.org/x/term"
)

var CreateUserCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new CastKeeper user",
	Long:  "Utility script for creating new users in the database for the given CastKeeper configuration.",
	Run:   run,
}

var (
	cfgFile         string
	username        string
	password        string
	accessLevel     int
	accessLevelHint string
)

func init() {
	hints := make([]string, 0)
	for _, al := range users.AccessLevels {
		hints = append(hints, fmt.Sprintf("%d = %s", al, al.Format()))
	}
	accessLevelHint = strings.Join(hints, ", ")

	accessLevelUsage := fmt.Sprintf("access level for the user to create (otherwise entered interactively): %s", accessLevelHint)

	CreateUserCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (otherwise uses default locations)")
	CreateUserCmd.Flags().StringVar(&username, "username", "", "username for the user to create (otherwise entered interactively)")
	CreateUserCmd.Flags().StringVar(&password, "password", "", "password for the user to create (otherwise entered interactively)")
	CreateUserCmd.Flags().IntVar(&accessLevel, "accessLevel", 0, accessLevelUsage)
}

func run(cmd *cobra.Command, args []string) {
	cfg, logger, err := config.LoadConfig(cfgFile)
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	ctx := framework.ContextWithLogger(context.Background(), logger)

	db, err := database.ConfigureDatabase(cfg, logger, false)
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

	accessLevel, err := readAccessLevel()
	if err != nil {
		log.Fatalf("failed to read access level: %v", err)
	}

	err = users.CreateUser(ctx, db, username, password, accessLevel)
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

func readAccessLevel() (users.AccessLevel, error) {
	argAccessLevel := accessLevel
	if argAccessLevel != 0 {
		return users.AccessLevel(argAccessLevel), nil
	}

	fmt.Printf("Enter access level (%s): ", accessLevelHint)
	accessLevelStr, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return 0, err
	}
	accessLevelStr = strings.TrimSpace(accessLevelStr)

	alInt, err := strconv.ParseInt(accessLevelStr, 10, 0)
	if err != nil {
		return 0, err
	}

	return users.AccessLevel(alInt), nil
}
