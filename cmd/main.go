package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/cmd/changepassword"
	"github.com/webbgeorge/castkeeper/cmd/createuser"
	"github.com/webbgeorge/castkeeper/cmd/deleteuser"
	"github.com/webbgeorge/castkeeper/cmd/edituser"
	"github.com/webbgeorge/castkeeper/cmd/listusers"
	"github.com/webbgeorge/castkeeper/cmd/serve"
	"github.com/webbgeorge/castkeeper/cmd/version"
)

func main() {
	userRootCmd := &cobra.Command{Use: "users"}
	userRootCmd.AddCommand(listusers.ListUsersCmd)
	userRootCmd.AddCommand(createuser.CreateUserCmd)
	userRootCmd.AddCommand(changepassword.ChangePasswordCmd)
	userRootCmd.AddCommand(edituser.EditUserCmd)
	userRootCmd.AddCommand(deleteuser.DeleteUserCmd)

	rootCmd := &cobra.Command{Use: "castkeeper"}
	rootCmd.AddCommand(
		serve.ServeCmd,
		userRootCmd,
		version.VersionCmd,
	)

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (otherwise uses default locations)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
