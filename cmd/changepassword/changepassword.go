package changepassword

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/auth/users"
	"github.com/webbgeorge/castkeeper/pkg/config/cli"
	"golang.org/x/term"
)

var ChangePasswordCmd = &cobra.Command{
	Use:   "change-password",
	Short: "Change the password of a user",
	Long:  "Utility script for changing user passwords in the database for the given CastKeeper configuration.",
	Run:   run,
}

var (
	username    string
	newPassword string
)

func init() {
	cli.InitGlobalFlags(ChangePasswordCmd)
	ChangePasswordCmd.Flags().StringVar(&username, "username", "", "username to change the password for")
	ChangePasswordCmd.Flags().StringVar(&newPassword, "new-password", "", "the new password to set for the user (otherwise entered interactively)")
	if err := ChangePasswordCmd.MarkFlagRequired("username"); err != nil {
		panic(err)
	}
}

func run(cmd *cobra.Command, args []string) {
	ctx, _, db, err := cli.ConfigureCLI()
	if err != nil {
		log.Fatal(err)
	}

	user, err := users.GetUserByUsername(ctx, db, username)
	if err != nil {
		log.Fatalf("failed to get user by username: %v", err)
	}

	password, err := readPassword()
	if err != nil {
		log.Fatalf("failed to read password: %v", err)
	}

	err = users.UpdatePassword(ctx, db, user.ID, password)
	if err != nil {
		log.Fatalf("failed to change password: %v", err)
	}

	log.Printf("successfully changed password for user '%s'", username)
}

func readPassword() (string, error) {
	argPassword := newPassword
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
