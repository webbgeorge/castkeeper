package listusers

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/auth/users"
	"github.com/webbgeorge/castkeeper/pkg/config/cli"
)

var ListUsersCmd = &cobra.Command{
	Use:   "list",
	Short: "List CastKeeper users",
	Long:  "Utility script for listing users from the database for the given CastKeeper configuration.",
	Run:   run,
}

func init() {
	cli.InitGlobalFlags(ListUsersCmd)
}

func run(cmd *cobra.Command, args []string) {
	ctx, _, db, err := cli.ConfigureCLI()
	if err != nil {
		log.Fatal(err)
	}

	userList, err := users.ListUsers(ctx, db)
	if err != nil {
		log.Fatalf("failed to list users: %v", err)
	}

	if len(userList) == 0 {
		fmt.Println("No users found")
		return
	}
	for _, user := range userList {
		fmt.Printf("%d\t%s (%s)\n", user.ID, user.Username, user.AccessLevel.Format())
	}
}
