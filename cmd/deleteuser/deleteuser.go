package deleteuser

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/auth/users"
	"github.com/webbgeorge/castkeeper/pkg/config/cli"
)

var DeleteUserCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a CastKeeper user",
	Long:  "Utility script for deleting users from the database for the given CastKeeper configuration.",
	Run:   run,
}

var username string

func init() {
	cli.InitGlobalFlags(DeleteUserCmd)
	DeleteUserCmd.Flags().StringVar(&username, "username", "", "username for the user to delete")
	if err := DeleteUserCmd.MarkFlagRequired("username"); err != nil {
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

	err = users.DeleteUser(ctx, db, user.ID)
	if err != nil {
		log.Fatalf("failed to delete user: %v", err)
	}

	log.Printf("successfully deleted user '%s'", username)
}
