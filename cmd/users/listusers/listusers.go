package listusers

import (
	"fmt"
	"log"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/auth/users"
	"github.com/webbgeorge/castkeeper/pkg/cli"
	cliconfig "github.com/webbgeorge/castkeeper/pkg/config/cli"
)

var ListUsersCmd = &cobra.Command{
	Use:   "list",
	Short: "List CastKeeper users",
	Long:  "Utility script for listing users from the database for the given CastKeeper configuration.",
	Run:   run,
}

func init() {
	cliconfig.InitGlobalFlags(ListUsersCmd)
}

func run(cmd *cobra.Command, args []string) {
	ctx, _, db, err := cliconfig.ConfigureCLI()
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

	rows := [][]string{
		{"ID", "Username", "Access Level"},
	}
	for _, user := range userList {
		rows = append(rows, []string{
			strconv.FormatInt(int64(user.ID), 10),
			user.Username,
			user.AccessLevel.Format(),
		})
	}
	cli.PrintTable(rows)
}
