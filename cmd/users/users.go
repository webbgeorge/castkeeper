package users

import (
	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/cmd/users/changepassword"
	"github.com/webbgeorge/castkeeper/cmd/users/createuser"
	"github.com/webbgeorge/castkeeper/cmd/users/deleteuser"
	"github.com/webbgeorge/castkeeper/cmd/users/edituser"
	"github.com/webbgeorge/castkeeper/cmd/users/listusers"
)

var Cmd = &cobra.Command{
	Use:     "users",
	Short:   "User management commands",
	GroupID: "commands",
}

func init() {
	Cmd.AddCommand(listusers.ListUsersCmd)
	Cmd.AddCommand(createuser.CreateUserCmd)
	Cmd.AddCommand(changepassword.ChangePasswordCmd)
	Cmd.AddCommand(edituser.EditUserCmd)
	Cmd.AddCommand(deleteuser.DeleteUserCmd)
}
