package edituser

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/auth/users"
	"github.com/webbgeorge/castkeeper/pkg/config/cli"
)

var EditUserCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a user",
	Long:  "Utility script for editing users in the database for the given CastKeeper configuration.",
	Run:   run,
}

var (
	username       string
	newUsername    string
	newAccessLevel int
)

func init() {
	cli.InitGlobalFlags(EditUserCmd)

	hints := make([]string, 0)
	for _, al := range users.AccessLevels {
		hints = append(hints, fmt.Sprintf("%d = %s", al, al.Format()))
	}
	accessLevelUsage := fmt.Sprintf("the new access level to set for the user (%s)", strings.Join(hints, ", "))

	EditUserCmd.Flags().StringVar(&username, "username", "", "username to identify the user to edit")
	EditUserCmd.Flags().StringVar(&newUsername, "new-username", "", "the new username to set for the user")
	EditUserCmd.Flags().IntVar(&newAccessLevel, "new-access-level", 0, accessLevelUsage)
	if err := EditUserCmd.MarkFlagRequired("username"); err != nil {
		panic(err)
	}
	if err := EditUserCmd.MarkFlagRequired("new-username"); err != nil {
		panic(err)
	}
	if err := EditUserCmd.MarkFlagRequired("new-access-level"); err != nil {
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

	err = users.UpdateUser(ctx, db, user.ID, newUsername, users.AccessLevel(newAccessLevel))
	if err != nil {
		log.Fatalf("failed to update user: %v", err)
	}

	log.Printf("successfully updated user '%s'", newUsername)
}
