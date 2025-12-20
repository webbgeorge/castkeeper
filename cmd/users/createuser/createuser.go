package createuser

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/auth/users"
	"github.com/webbgeorge/castkeeper/pkg/config/cli"
	"golang.org/x/term"
)

var CreateUserCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new CastKeeper user",
	Long:  "Utility script for creating new users in the database for the given CastKeeper configuration.",
	Run:   run,
}

var (
	username    string
	password    string
	accessLevel int
)

func init() {
	cli.InitGlobalFlags(CreateUserCmd)

	hints := make([]string, 0)
	for _, al := range users.AccessLevels {
		hints = append(hints, fmt.Sprintf("%d = %s", al, al.Format()))
	}
	accessLevelUsage := fmt.Sprintf("access level for the user to create (%s)", strings.Join(hints, ", "))

	CreateUserCmd.Flags().StringVar(&username, "username", "", "username for the user to create")
	CreateUserCmd.Flags().StringVar(&password, "password", "", "password for the user to create (otherwise entered interactively)")
	CreateUserCmd.Flags().IntVar(&accessLevel, "access-level", 0, accessLevelUsage)
	if err := CreateUserCmd.MarkFlagRequired("username"); err != nil {
		panic(err)
	}
	if err := CreateUserCmd.MarkFlagRequired("access-level"); err != nil {
		panic(err)
	}
}

func run(cmd *cobra.Command, args []string) {
	ctx, _, db, err := cli.ConfigureCLI()
	if err != nil {
		log.Fatal(err)
	}

	password, err := readPassword()
	if err != nil {
		log.Fatalf("failed to read password: %v", err)
	}

	err = users.CreateUser(ctx, db, username, password, users.AccessLevel(accessLevel))
	if err != nil {
		log.Fatalf("failed to create user: %v", err)
	}

	log.Printf("successfully created user '%s'", username)
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
